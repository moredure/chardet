package chardet

import "sync"

type recognizer interface {
	Match(*recognizerInput) recognizerOutput
}

type recognizerOutput Result

type recognizerInput struct {
	raw         []byte
	input       []byte
	tagStripped bool
	hasC1Bytes  bool
}

const inputBufferSize = 8192

var recognizerInputsPool = sync.Pool{
	New: func() interface{} {
		return &recognizerInput{
			input: make([]byte, 0, inputBufferSize),
		}
	},
}


func (r *recognizerInput) Reset() {
	r.raw = nil
}

func newRecognizerInput(raw []byte, stripTag bool) *recognizerInput {
	input := recognizerInputsPool.Get().(*recognizerInput)
	input.input, input.tagStripped = mayStripInput(input.input[:0], raw, stripTag)
	input.hasC1Bytes = computeHasC1Bytes(input.input)
	input.raw = raw
	return input
}


func mayStripInput(input, raw []byte, stripTag bool) (out []byte, stripped bool) {
	out = input
	var badTags, openTags int32
	inMarkup := false
	stripped = false
	if stripTag {
		stripped = true
		for _, c := range raw {
			if c == '<' {
				if inMarkup {
					badTags += 1
				}
				inMarkup = true
				openTags += 1
			}
			if !inMarkup {
				out = append(out, c)
				if len(out) >= inputBufferSize {
					break
				}
			}
			if c == '>' {
				inMarkup = false
			}
		}
	}
	if openTags < 5 || openTags/5 < badTags || (len(out) < 100 && len(raw) > 600) {
		limit := len(raw)
		if limit > inputBufferSize {
			limit = inputBufferSize
		}
		out = out[:limit]
		copy(out, raw[:limit])
		stripped = false
	}
	return
}

func computeHasC1Bytes(input []byte) bool {
	var byteStats [256]byte
	for _, c := range input {
		byteStats[c] += 1
	}
	for _, count := range byteStats[0x80 : 0x9F+1] {
		if count > 0 {
			return true
		}
	}
	return false
}
