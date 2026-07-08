// SPDX-License-Identifier: MIT
package app

import (
	"encoding/binary"
	"fmt"
)

const (
	zipLocalHeaderSig     = 0x04034b50
	zipCentralHeaderSig   = 0x02014b50
	zipEndCentralSig      = 0x06054b50
	zipLocalHeaderMinSize = 30
	zipCentralHeaderSize  = 46
	zipEndCentralSize     = 22
)

type zipLocalEntry struct {
	offset           uint32
	name             string
	method           uint16
	flags            uint16
	crc32            uint32
	compressedSize   uint32
	uncompressedSize uint32
	modTime          uint16
	modDate          uint16
	extra            []byte
}

func zipHasEndOfCentralDirectory(body []byte) bool {
	if len(body) < zipEndCentralSize {
		return false
	}
	searchStart := len(body) - zipEndCentralSize
	minStart := len(body) - 65557
	if minStart < 0 {
		minStart = 0
	}
	for i := searchStart; i >= minStart; i-- {
		if binary.LittleEndian.Uint32(body[i:]) == zipEndCentralSig {
			return true
		}
	}
	return false
}

func scanLocalZipEntries(body []byte) ([]zipLocalEntry, int, error) {
	if len(body) < zipLocalHeaderMinSize {
		return nil, 0, fmt.Errorf("epub file is too small or incomplete")
	}
	if body[0] != 'P' || body[1] != 'K' || body[2] != 3 || body[3] != 4 {
		return nil, 0, fmt.Errorf("file is not a valid epub (missing zip header)")
	}

	var entries []zipLocalEntry
	offset := 0
	for offset+zipLocalHeaderMinSize <= len(body) {
		if binary.LittleEndian.Uint32(body[offset:]) != zipLocalHeaderSig {
			if len(entries) == 0 {
				return nil, 0, fmt.Errorf("epub zip structure is invalid")
			}
			return entries, offset, nil
		}
		flags := binary.LittleEndian.Uint16(body[offset+6:])
		method := binary.LittleEndian.Uint16(body[offset+8:])
		modTime := binary.LittleEndian.Uint16(body[offset+10:])
		modDate := binary.LittleEndian.Uint16(body[offset+12:])
		crc32 := binary.LittleEndian.Uint32(body[offset+14:])
		compressedSize := binary.LittleEndian.Uint32(body[offset+18:])
		uncompressedSize := binary.LittleEndian.Uint32(body[offset+22:])
		nameLen := int(binary.LittleEndian.Uint16(body[offset+26:]))
		extraLen := int(binary.LittleEndian.Uint16(body[offset+28:]))
		nameStart := offset + zipLocalHeaderMinSize
		nameEnd := nameStart + nameLen
		extraEnd := nameEnd + extraLen
		dataStart := extraEnd
		if nameEnd > len(body) || extraEnd > len(body) || dataStart > len(body) {
			if len(entries) == 0 {
				return nil, 0, fmt.Errorf("epub zip structure is invalid")
			}
			return entries, offset, nil
		}
		if flags&0x8 != 0 {
			return nil, 0, fmt.Errorf("epub zip uses streaming descriptors and cannot be repaired")
		}
		name := string(body[nameStart:nameEnd])
		if name == "" {
			return nil, 0, fmt.Errorf("epub zip structure is invalid")
		}
		extra := append([]byte(nil), body[nameEnd:extraEnd]...)
		dataEnd := dataStart + int(compressedSize)
		if dataEnd > len(body) {
			if len(entries) == 0 {
				return nil, 0, fmt.Errorf("epub zip structure is invalid")
			}
			return entries, offset, nil
		}
		entries = append(entries, zipLocalEntry{
			offset:           uint32(offset),
			name:             name,
			method:           method,
			flags:            flags,
			crc32:            crc32,
			compressedSize:   compressedSize,
			uncompressedSize: uncompressedSize,
			modTime:          modTime,
			modDate:          modDate,
			extra:            extra,
		})
		offset = dataEnd
	}
	if len(entries) == 0 {
		return nil, 0, fmt.Errorf("epub zip has no entries")
	}
	return entries, offset, nil
}

func u16ZipField(n int, label string) (uint16, error) {
	if n < 0 || n > 0xffff {
		return 0, fmt.Errorf("zip %s exceeds uint16", label)
	}
	return uint16(n), nil
}

func u32ZipField(n int, label string) (uint32, error) {
	if n < 0 || int64(n) > 0xffffffff {
		return 0, fmt.Errorf("zip %s exceeds uint32", label)
	}
	return uint32(n), nil
}

func appendZipCentralDirectory(body []byte, entries []zipLocalEntry) ([]byte, error) {
	cdStart := len(body)
	out := append([]byte(nil), body...)
	for _, entry := range entries {
		nameBytes := []byte(entry.name)
		nameLen, err := u16ZipField(len(nameBytes), "entry name length")
		if err != nil {
			return nil, err
		}
		extraLen, err := u16ZipField(len(entry.extra), "entry extra length")
		if err != nil {
			return nil, err
		}
		header := make([]byte, zipCentralHeaderSize+len(nameBytes)+len(entry.extra))
		binary.LittleEndian.PutUint32(header[0:], zipCentralHeaderSig)
		binary.LittleEndian.PutUint16(header[4:], 20)
		binary.LittleEndian.PutUint16(header[6:], 20)
		binary.LittleEndian.PutUint16(header[8:], entry.flags)
		binary.LittleEndian.PutUint16(header[10:], entry.method)
		binary.LittleEndian.PutUint16(header[12:], entry.modTime)
		binary.LittleEndian.PutUint16(header[14:], entry.modDate)
		binary.LittleEndian.PutUint32(header[16:], entry.crc32)
		binary.LittleEndian.PutUint32(header[20:], entry.compressedSize)
		binary.LittleEndian.PutUint32(header[24:], entry.uncompressedSize)
		binary.LittleEndian.PutUint16(header[28:], nameLen)
		binary.LittleEndian.PutUint16(header[30:], extraLen)
		binary.LittleEndian.PutUint16(header[32:], 0)
		binary.LittleEndian.PutUint16(header[34:], 0)
		binary.LittleEndian.PutUint16(header[36:], 0)
		binary.LittleEndian.PutUint32(header[38:], 0)
		binary.LittleEndian.PutUint32(header[42:], entry.offset)
		copy(header[zipCentralHeaderSize:], nameBytes)
		copy(header[zipCentralHeaderSize+len(nameBytes):], entry.extra)
		out = append(out, header...)
	}
	cdSize := len(out) - cdStart
	entryCount, err := u16ZipField(len(entries), "entry count")
	if err != nil {
		return nil, err
	}
	cdSizeU32, err := u32ZipField(cdSize, "central directory size")
	if err != nil {
		return nil, err
	}
	cdStartU32, err := u32ZipField(cdStart, "central directory offset")
	if err != nil {
		return nil, err
	}
	eocd := make([]byte, zipEndCentralSize)
	binary.LittleEndian.PutUint32(eocd[0:], zipEndCentralSig)
	binary.LittleEndian.PutUint16(eocd[8:], entryCount)
	binary.LittleEndian.PutUint16(eocd[10:], entryCount)
	binary.LittleEndian.PutUint32(eocd[12:], cdSizeU32)
	binary.LittleEndian.PutUint32(eocd[16:], cdStartU32)
	out = append(out, eocd...)
	return out, nil
}

func repairZipIfNeeded(body []byte) ([]byte, error) {
	if zipHasEndOfCentralDirectory(body) {
		return body, nil
	}
	entries, endOffset, err := scanLocalZipEntries(body)
	if err != nil {
		return nil, err
	}
	trimmed := body[:endOffset]
	repaired, err := appendZipCentralDirectory(trimmed, entries)
	if err != nil {
		return nil, err
	}
	if !zipHasEndOfCentralDirectory(repaired) {
		return nil, fmt.Errorf("epub zip repair failed")
	}
	return repaired, nil
}
