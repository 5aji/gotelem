// Code generated by "stringer -output=frame_kind.go -type Kind"; DO NOT EDIT.

package gotelem

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[CanSFFFrame-0]
	_ = x[CanEFFFrame-1]
	_ = x[CanRTRFrame-2]
	_ = x[CanErrFrame-3]
}

const _Kind_name = "SFFEFFRTRERR"

var _Kind_index = [...]uint8{0, 3, 6, 9, 12}

func (i Kind) String() string {
	if i >= Kind(len(_Kind_index)-1) {
		return "Kind(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Kind_name[_Kind_index[i]:_Kind_index[i+1]]
}
