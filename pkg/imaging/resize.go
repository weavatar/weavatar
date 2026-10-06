package imaging

import (
	"image"
	"image/draw"
	"math"
)

// Resize scales bilinearly, widening the kernel by the scale when shrinking so
// every source pixel contributes.
func Resize(img image.Image, width, height int) image.Image {
	b := img.Bounds()
	if b.Dx() == width && b.Dy() == height {
		return img
	}

	xw := newWeights(b.Dx(), width)
	yw := newWeights(b.Dy(), height)
	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	// rows are converted to premultiplied RGBA so interpolation does not bleed
	// color from transparent pixels
	row := image.NewRGBA(image.Rect(0, 0, b.Dx(), 1))
	// horizontally scaled rows, reused modulo the vertical window size
	ring := make([][]float32, yw.span)
	for i := range ring {
		ring[i] = make([]float32, width*4)
	}
	acc := make([]float32, width*4)
	next := 0 // next source row to scale horizontally

	for y := range height {
		lo, ws := yw.start[y], yw.at(y)
		for ; next < lo+len(ws); next++ {
			if next < lo {
				continue
			}
			draw.Draw(row, row.Rect, img, image.Pt(b.Min.X, b.Min.Y+next), draw.Src)
			xw.apply(ring[next%yw.span], row.Pix)
		}

		clear(acc)
		for k, v := range ws {
			for i, s := range ring[(lo+k)%yw.span] {
				acc[i] += v * s
			}
		}
		out := dst.Pix[y*dst.Stride : y*dst.Stride+width*4]
		for i, s := range acc {
			out[i] = uint8(min(max(s+0.5, 0), 255))
		}
	}

	return dst
}

// weights holds the kernel along one axis: destination pixel i takes weights
// w[i*stride : i*stride+n[i]] over source pixels from start[i].
type weights struct {
	start  []int
	n      []int
	w      []float32 // normalized
	stride int
	span   int // widest window
}

func newWeights(src, dst int) weights {
	scale := float64(src) / float64(dst)
	support := max(scale, 1)
	stride := int(math.Ceil(2*support)) + 1

	ws := weights{
		start:  make([]int, dst),
		n:      make([]int, dst),
		w:      make([]float32, dst*stride),
		stride: stride,
	}
	buf := make([]float64, stride)
	for i := range dst {
		center := (float64(i)+0.5)*scale - 0.5
		lo := max(int(math.Floor(center-support)), 0)
		hi := min(int(math.Ceil(center+support)), src)

		var sum float64
		tmp := buf[:hi-lo]
		for j := lo; j < hi; j++ {
			v := max(1-math.Abs(center-float64(j))/support, 0)
			tmp[j-lo] = v
			sum += v
		}
		// trim zero weights at both ends
		for len(tmp) > 0 && tmp[0] == 0 {
			tmp, lo = tmp[1:], lo+1
		}
		for len(tmp) > 0 && tmp[len(tmp)-1] == 0 {
			tmp = tmp[:len(tmp)-1]
		}

		ws.start[i], ws.n[i] = lo, len(tmp)
		for k, v := range tmp {
			ws.w[i*stride+k] = float32(v / sum)
		}
		ws.span = max(ws.span, len(tmp))
	}

	return ws
}

func (ws weights) at(i int) []float32 {
	return ws.w[i*ws.stride : i*ws.stride+ws.n[i]]
}

func (ws weights) apply(dst []float32, src []uint8) {
	for i := range ws.start {
		var r, g, b, a float32
		s := src[ws.start[i]*4:]
		for k, v := range ws.at(i) {
			p := s[k*4 : k*4+4 : k*4+4]
			r += v * float32(p[0])
			g += v * float32(p[1])
			b += v * float32(p[2])
			a += v * float32(p[3])
		}
		d := dst[i*4 : i*4+4 : i*4+4]
		d[0], d[1], d[2], d[3] = r, g, b, a
	}
}
