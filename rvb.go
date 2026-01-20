package rvb

import "sync"

type Checkpoint struct {
	rotation int
	pos      int
}

type RVBuffer[E any] struct {
	buffer []E
	size   int

	mtx             sync.RWMutex
	currentPos      int
	currentSize     int
	currentRotation int
}

func NewBuffer[E any](size int) *RVBuffer[E] {
	return &RVBuffer[E]{
		buffer:          make([]E, size),
		size:            size,
		currentPos:      0,
		currentSize:     0,
		currentRotation: 0,
		mtx:             sync.RWMutex{},
	}
}

// GetCurrentSIze will return how many elements is currently being stored in the buffer,
// up to `size` defined at the buffer creation.
func (b *RVBuffer[E]) GetCurrentSIze() int {
	return b.currentSize
}

// Push will remove the oldest items if maximum buffer size has been reached
func (b *RVBuffer[E]) Push(item E) {
	b.mtx.Lock()
	b.buffer[b.currentPos] = item
	b.currentSize = min(b.size, max(b.currentPos+1, b.currentSize))
	if b.currentPos+1 >= b.size {
		b.currentRotation++
	}
	b.currentPos = (b.currentPos + 1) % b.size
	b.mtx.Unlock()
}

// PushMany will remove the oldest items if maximum buffer size has been reached
func (b *RVBuffer[E]) PushMany(items []E) {
	b.mtx.Lock()
	for _, item := range items {
		b.buffer[b.currentPos] = item
		b.currentSize = min(b.size, max(b.currentPos+1, b.currentSize))
		if b.currentPos+1 >= b.size {
			b.currentRotation++
		}
		b.currentPos = (b.currentPos + 1) % b.size
	}
	b.mtx.Unlock()
}

// ReadNew will return up to `n` the newest items, from the newest to the oldest.
// Size of returned slice will be lower than `n` when not enough items is available in the buffer to fulfill request.
func (b *RVBuffer[E]) ReadNew(n int) (out []E) {
	b.mtx.RLock()
	out = make([]E, min(n, b.currentSize))
	for i := 0; i < min(n, b.currentSize); i++ {
		out[i] = b.buffer[(b.currentSize+b.currentPos-1-i)%b.currentSize]
	}
	b.mtx.RUnlock()
	return out
}

// ReadOld will return up to `n` the oldest items, from the oldest to the newest.
// Size of returned slice will be lower than `n` when not enough items is available in the buffer to fulfill request.
func (b *RVBuffer[E]) ReadOld(n int) (out []E) {
	b.mtx.RLock()
	out = make([]E, min(n, b.currentSize))
	for i := 0; i < min(n, b.currentSize); i++ {
		out[i] = b.buffer[(b.currentPos+i)%b.currentSize]
	}
	b.mtx.RUnlock()
	return out
}

// GetCheckpoint will return a magic value (static location) that will let user efficiently read the newest messages from the buffer.
// It is mainly focused to fetch buffer history with ReadNewFromCheckpoint in small read requests to avoid blocking the buffer of item pusher.
func (b *RVBuffer[E]) GetCheckpoint() Checkpoint {
	b.mtx.RLock()
	cp := Checkpoint{
		pos:      b.currentPos,
		rotation: b.currentRotation,
	}
	b.mtx.RUnlock()
	return cp
}

// NewItemsSince will return the number of new items that appeared since given checkpoint.
func (b *RVBuffer[E]) NewItemsSince(checkpoint Checkpoint) int {
	b.mtx.RLock()
	currentAbs := b.currentRotation*b.size + b.currentPos
	checkpointAbs := checkpoint.rotation*b.size + checkpoint.pos
	b.mtx.RUnlock()

	if checkpointAbs > currentAbs {
		return 0
	}
	return currentAbs - checkpointAbs
}

type Missing struct {
	// Reused reports how many items has been already overwritten and is no longer available from request point of view
	Reused int
	// Max reports how many items could be returned without considering buffer overwriting from request point of view
	Max int
}

// ReadNewFromCheckpoint will return up to `n` the newest items from a given `checkpoint` (static location),
// from the newest to the oldest, with skipped `skip` items (useful for pagination or scrolling).
// Size of returned slice will be lower than `n` when not enough items is available in the buffer to fulfill request,
// either because there was too few items available or because some items has been already overwritten internally.
// Additionally, `missing` will be returned as second value that will be helpful to recognize overwriting situation,
func (b *RVBuffer[E]) ReadNewFromCheckpoint(checkpoint Checkpoint, skip, n int) (out []E, missing Missing) {
	b.mtx.RLock()

	currentAbs := b.currentRotation*b.size + b.currentPos
	checkpointAbs := checkpoint.rotation*b.size + checkpoint.pos

	availableAtCheckpoint := b.size
	if checkpoint.rotation == 0 {
		availableAtCheckpoint = checkpoint.pos
	}

	oldestValidAbs := currentAbs - b.currentSize
	startReadAbs := checkpointAbs - 1 - skip

	effectiveAvailable := max(0, availableAtCheckpoint-skip)
	itemsToRead := min(n, effectiveAvailable)

	out = make([]E, 0, itemsToRead)
	reusedCount := 0

	for i := 0; i < itemsToRead; i++ {
		targetAbs := startReadAbs - i

		if targetAbs < oldestValidAbs {
			reusedCount++
		} else {
			bufferIdx := targetAbs % b.size
			out = append(out, b.buffer[bufferIdx])
		}
	}

	b.mtx.RUnlock()
	return out, Missing{
		Reused: reusedCount,
		Max:    itemsToRead,
	}
}
