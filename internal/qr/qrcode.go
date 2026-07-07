// SPDX-License-Identifier: MIT

// Package qr implements a minimal QR Code Model 2 encoder (byte mode, ECC-M).
package qr

import (
	"encoding/base64"
	"fmt"
	"strings"
)

const maskPattern = 0

type versionSpec struct {
	size      int
	dataBytes int
	ecPerBlk  int
	blocks    int
}

// EC-M byte-mode capacities for versions 1-10.
var versions = [...]versionSpec{
	{21, 14, 10, 1},
	{25, 26, 16, 1},
	{29, 42, 26, 1},
	{33, 62, 18, 2},
	{37, 84, 18, 2},
	{41, 106, 18, 2},
	{45, 122, 20, 2},
	{49, 152, 24, 2},
	{53, 180, 24, 2},
	{57, 213, 26, 2},
}

var alignCoords = [][]int{
	nil,
	{6, 18},
	{6, 22},
	{6, 26},
	{6, 30},
	{6, 34},
	{6, 22, 38},
	{6, 24, 42},
	{6, 26, 46},
	{6, 28, 50},
}

var formatBitsMedium = [8]uint16{
	0x5412, 0x5115, 0x5e7c, 0x5b4b, 0x45f9, 0x40ce, 0x4f97, 0x4aa0,
}

var versionBits = [34]uint32{
	0x07c94, 0x085bc, 0x09a99, 0x0a4d3, 0x0bbf6, 0x0c762, 0x0d847, 0x0e60d,
	0x0f928, 0x10b78, 0x1145d, 0x12a17, 0x13532, 0x149a6, 0x15683, 0x168c9,
	0x177ec, 0x18ec4, 0x191e1, 0x1afab, 0x1b08e, 0x1cc1a, 0x1d33f, 0x1ed75,
	0x1f250, 0x209d5, 0x216f0, 0x228ba, 0x2379f, 0x24b0b, 0x2542e, 0x26a64,
	0x27541, 0x28c69,
}

type gf256 struct {
	exp [512]byte
	log [256]byte
}

func newGF256() gf256 {
	var g gf256
	x := 1
	for i := 0; i < 255; i++ {
		g.exp[i] = byte(x)
		g.log[x] = byte(i)
		x <<= 1
		if x&0x100 != 0 {
			x ^= 0x11D
		}
	}
	for i := 255; i < 512; i++ {
		g.exp[i] = g.exp[i-255]
	}
	return g
}

var gf = newGF256()

func gfMul(a, b byte) byte {
	if a == 0 || b == 0 {
		return 0
	}
	return gf.exp[int(gf.log[a])+int(gf.log[b])]
}

func rsEncode(data []byte, ecCount int) []byte {
	gen := make([]byte, ecCount+1)
	gen[0] = 1
	for i := 0; i < ecCount; i++ {
		for j := ecCount; j > 0; j-- {
			gen[j] = gfMul(gen[j], gf.exp[i]) ^ gen[j-1]
		}
		gen[0] = gfMul(gen[0], gf.exp[i])
	}
	msg := append(append([]byte{}, data...), make([]byte, ecCount)...)
	for i := 0; i < len(data); i++ {
		coef := msg[i]
		if coef == 0 {
			continue
		}
		for j := 0; j <= ecCount; j++ {
			msg[i+j] ^= gfMul(gen[j], coef)
		}
	}
	return msg[len(data):]
}

type bitBuffer struct {
	bits []bool
}

func (b *bitBuffer) write(val, n int) {
	for i := n - 1; i >= 0; i-- {
		b.bits = append(b.bits, (val>>uint(i))&1 == 1)
	}
}

func (b *bitBuffer) codewords(n int) []byte {
	out := make([]byte, n)
	for i := 0; i < len(b.bits) && i < n*8; i++ {
		if b.bits[i] {
			out[i/8] |= 1 << uint(7-i%8)
		}
	}
	return out
}

func countIndicatorBits(ver int) int {
	if ver < 10 {
		return 8
	}
	return 16
}

func encodeDataCodewords(data []byte, ver int) []byte {
	v := versions[ver-1]
	var buf bitBuffer
	buf.write(4, 4)
	buf.write(len(data), countIndicatorBits(ver))
	for _, c := range data {
		buf.write(int(c), 8)
	}
	term := v.dataBytes*8 - len(buf.bits)
	if term > 4 {
		term = 4
	}
	if term > 0 {
		buf.write(0, term)
	}
	for len(buf.bits)%8 != 0 {
		buf.write(0, 1)
	}
	raw := buf.codewords(v.dataBytes)
	for len(raw) < v.dataBytes {
		if len(raw)%2 == 0 {
			raw = append(raw, 0xEC)
		} else {
			raw = append(raw, 0x11)
		}
	}
	return raw[:v.dataBytes]
}

func interleave(ver int, data []byte) []byte {
	v := versions[ver-1]
	blockData := v.dataBytes / v.blocks
	ec := v.ecPerBlk
	blocks := make([][]byte, v.blocks)
	for i := 0; i < v.blocks; i++ {
		start := i * blockData
		chunk := append([]byte{}, data[start:start+blockData]...)
		blocks[i] = append(chunk, rsEncode(chunk, ec)...)
	}
	out := make([]byte, 0, (blockData+ec)*v.blocks)
	for offset := 0; offset < blockData; offset++ {
		for i := 0; i < v.blocks; i++ {
			out = append(out, blocks[i][offset])
		}
	}
	for offset := 0; offset < ec; offset++ {
		for i := 0; i < v.blocks; i++ {
			out = append(out, blocks[i][blockData+offset])
		}
	}
	return out
}

const (
	modBlank = 0
	modDark  = 1
	modFunc  = 2
)

type matrix struct {
	size int
	cell []byte
}

func newMatrix(size int) *matrix {
	return &matrix{size: size, cell: make([]byte, size*size)}
}

func (m *matrix) get(x, y int) byte {
	if x < 0 || y < 0 || x >= m.size || y >= m.size {
		return modFunc
	}
	return m.cell[y*m.size+x]
}

func (m *matrix) set(x, y int, v byte) {
	if x < 0 || y < 0 || x >= m.size || y >= m.size {
		return
	}
	m.cell[y*m.size+x] = v
}

func (m *matrix) setFunc(x, y int, dark bool) {
	v := byte(modBlank)
	if dark {
		v = modDark
	}
	m.set(x, y, modFunc|v)
}

func (m *matrix) isDark(x, y int) bool {
	return m.get(x, y)&modDark != 0
}

func placeFinder(m *matrix, x, y int) {
	for dy := -1; dy <= 7; dy++ {
		for dx := -1; dx <= 7; dx++ {
			xx, yy := x+dx, y+dy
			switch {
			case dx == -1 || dx == 7 || dy == -1 || dy == 7:
				m.setFunc(xx, yy, false)
			case dx == 0 || dx == 6 || dy == 0 || dy == 6:
				m.setFunc(xx, yy, true)
			case dx >= 2 && dx <= 4 && dy >= 2 && dy <= 4:
				m.setFunc(xx, yy, true)
			default:
				m.setFunc(xx, yy, false)
			}
		}
	}
}

func placeAlignment(m *matrix, cx, cy int) {
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			dark := dx == 0 || dy == 0 || (abs(dx) == 2 && abs(dy) == 2)
			m.setFunc(cx+dx, cy+dy, dark)
		}
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func placeTiming(m *matrix) {
	for i := 8; i < m.size-8; i++ {
		dark := i%2 == 0
		if m.get(i, 6) == modBlank {
			m.setFunc(i, 6, dark)
		}
		if m.get(6, i) == modBlank {
			m.setFunc(6, i, dark)
		}
	}
}

func reserveMetadata(m *matrix, ver int) {
	size := m.size
	for i := 0; i < 8; i++ {
		if i != 6 {
			m.setFunc(8, i, false)
		}
		m.setFunc(i, 8, false)
	}
	for i := 0; i < 7; i++ {
		m.setFunc(8, size-1-i, false)
		m.setFunc(size-1-i, 8, false)
	}
	m.setFunc(8, 8, false)
	m.setFunc(8, size-8, true)
	if ver >= 7 {
		for i := 0; i < 6; i++ {
			m.setFunc(8, size-11+i, false)
			m.setFunc(size-11+i, 8, false)
		}
	}
}

func placeFormat(m *matrix, bits uint16) {
	coords := [15][2]int{
		{8, 0}, {8, 1}, {8, 2}, {8, 3}, {8, 4}, {8, 5}, {8, 7}, {8, 8},
		{7, 8}, {5, 8}, {4, 8}, {3, 8}, {2, 8}, {1, 8}, {0, 8},
	}
	for i, c := range coords {
		m.setFunc(c[0], c[1], (bits>>uint(14-i))&1 == 1)
	}
	for i := 0; i < 8; i++ {
		m.setFunc(m.size-1-i, 8, (bits>>uint(i))&1 == 1)
	}
	for i := 0; i < 7; i++ {
		m.setFunc(8, m.size-7+i, (bits>>uint(8+i))&1 == 1)
	}
}

func placeVersion(m *matrix, ver int, bits uint32) {
	if ver < 7 {
		return
	}
	coords := [18][2]int{
		{0, m.size - 11}, {1, m.size - 11}, {2, m.size - 11},
		{0, m.size - 10}, {1, m.size - 10}, {2, m.size - 10},
		{0, m.size - 9}, {1, m.size - 9}, {2, m.size - 9},
		{m.size - 11, 0}, {m.size - 10, 0}, {m.size - 9, 0},
		{m.size - 11, 1}, {m.size - 10, 1}, {m.size - 9, 1},
		{m.size - 11, 2}, {m.size - 10, 2}, {m.size - 9, 2},
	}
	for i, c := range coords {
		m.setFunc(c[0], c[1], (bits>>uint(17-i))&1 == 1)
	}
}

func placeData(m *matrix, codewords []byte) {
	bitIdx := 0
	up := true
	for col := m.size - 1; col > 0; col -= 2 {
		if col == 6 {
			col--
		}
		for row := 0; row < m.size; row++ {
			y := row
			if !up {
				y = m.size - 1 - row
			}
			for dx := 0; dx < 2; dx++ {
				x := col - dx
				if m.get(x, y)&modFunc != 0 {
					continue
				}
				dark := false
				if bitIdx/8 < len(codewords) {
					dark = (codewords[bitIdx/8]>>(7-bitIdx%8))&1 == 1
				}
				if dark {
					m.set(x, y, modDark)
				} else {
					m.set(x, y, modBlank)
				}
				bitIdx++
			}
		}
		up = !up
	}
}

func maskFlip(x, y int) bool {
	return (x+y)%2 == 0
}

func applyMask(m *matrix) {
	for y := 0; y < m.size; y++ {
		for x := 0; x < m.size; x++ {
			if m.get(x, y)&modFunc != 0 {
				continue
			}
			if !maskFlip(x, y) {
				continue
			}
			if m.isDark(x, y) {
				m.set(x, y, modBlank)
			} else {
				m.set(x, y, modDark)
			}
		}
	}
}

func buildMatrix(ver int, codewords []byte) *matrix {
	size := versions[ver-1].size
	m := newMatrix(size)
	placeFinder(m, 0, 0)
	placeFinder(m, size-7, 0)
	placeFinder(m, 0, size-7)
	for _, cy := range alignCoords[ver-1] {
		for _, cx := range alignCoords[ver-1] {
			if (cx == 6 && cy == 6) || (cx == 6 && cy == size-7) || (cx == size-7 && cy == 6) {
				continue
			}
			placeAlignment(m, cx, cy)
		}
	}
	placeTiming(m)
	reserveMetadata(m, ver)
	placeData(m, codewords)
	applyMask(m)
	placeFormat(m, formatBitsMedium[maskPattern])
	if ver >= 7 {
		placeVersion(m, ver, versionBits[ver-7])
	}
	return m
}

func encodeMatrix(text string) (*matrix, error) {
	data := []byte(strings.TrimSpace(text))
	if len(data) == 0 {
		return nil, fmt.Errorf("qr: empty text")
	}
	for ver := 1; ver <= len(versions); ver++ {
		need := 4 + countIndicatorBits(ver) + len(data)*8 + 4
		if (need+7)/8 > versions[ver-1].dataBytes {
			continue
		}
		raw := encodeDataCodewords(data, ver)
		codewords := interleave(ver, raw)
		return buildMatrix(ver, codewords), nil
	}
	return nil, fmt.Errorf("qr: payload too large (%d bytes)", len(data))
}

// EncodeSVG returns a square SVG for text using byte mode and ECC-M.
func EncodeSVG(text string, moduleScale int) (string, error) {
	if moduleScale < 1 {
		moduleScale = 4
	}
	m, err := encodeMatrix(text)
	if err != nil {
		return "", err
	}
	return matrixSVG(m, moduleScale), nil
}

func matrixSVG(m *matrix, scale int) string {
	qz := 4 * scale
	size := m.size * scale
	total := size + 2*qz
	var b strings.Builder
	b.Grow(total * total / 2)
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d"><rect width="100%%" height="100%%" fill="#fff"/>`, total, total, total, total)
	for y := 0; y < m.size; y++ {
		for x := 0; x < m.size; x++ {
			if !m.isDark(x, y) {
				continue
			}
			px := qz + x*scale
			py := qz + y*scale
			fmt.Fprintf(&b, `<rect x="%d" y="%d" width="%d" height="%d" fill="#000"/>`, px, py, scale, scale)
		}
	}
	b.WriteString("</svg>")
	return b.String()
}

// DataURL returns a base64 SVG data URL suitable for <img src>.
func DataURL(text string, moduleScale int) (string, error) {
	svg, err := EncodeSVG(text, moduleScale)
	if err != nil {
		return "", err
	}
	return "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg)), nil
}

// Size returns the module count per side for encoded text.
func Size(text string) (int, error) {
	m, err := encodeMatrix(text)
	if err != nil {
		return 0, err
	}
	return m.size, nil
}
