// Niels Widger
// Time-stamp: <30 Jul 2013 at 18:25:08 by nwidger on macros.local>

package gotcl

import (
	"container/list"
	"errors"
)

type Stack struct {
	level_list *list.List
	level_map  map[int]*Frame
}

func (stack *Stack) PushFrame() *Frame {
	frame := NewFrame()
	top := 0

	if stack.level_list.Len() != 0 {
		top = stack.level_list.Front().Value.(*Frame).level
	}

	frame.level = top + 1
	stack.level_map[frame.level] = frame
	stack.level_list.PushFront(frame)

	return frame
}

func (stack *Stack) GetFrame(level int) (frame *Frame, error error) {
	frame, ok := stack.level_map[level]

	if !ok {
		return nil, errors.New("No frame at that level")
	}

	return frame, nil
}

func (stack *Stack) PeekFrame() *Frame {
	if stack.level_list.Len() != 0 {
		front := stack.level_list.Front()
		frame := front.Value.(*Frame)
		return frame
	}

	return nil
}

func (stack *Stack) PopFrame() {
	if stack.level_list.Len() != 0 {
		front := stack.level_list.Front()
		frame := front.Value.(*Frame)
		stack.level_map[frame.level] = nil
		stack.level_list.Remove(front)
	}
}
