package stacktrace

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/core/vm"
)

//@TODO: Use a more memory efficient method of representing source maps. How costly are InstructionMapping refs?

type InstructionMapping struct {
	Start  int
	Length int

	// WHAT WE REALLY NEED
	Line   int
	Column int

	FileIndex int
	Jump      string
}

// SourceMap is the memory address to instruction information map.
type SourceMap map[int]*InstructionMapping

func ParseSourceMap(sourceMap string, source string, bytecode string) (*SourceMap, error) {
	instructionSrcMap := make(SourceMap)

	var err error
	var s int64
	var l int64
	var f int64
	var j string
	for index, mapping := range strings.Split(sourceMap, ";") {
		if mapping == "" {
			instructionSrcMap[index] = instructionSrcMap[index-1]
			continue
		}

		info := strings.Split(mapping, ":")
		if len(info) > 0 && info[0] != "" {
			s, err = strconv.ParseInt(info[0], 0, 64)
			if err != nil {
				return nil, fmt.Errorf("failed parsing integer: %s", err)
			}
		}
		if len(info) > 1 && info[1] != "" {
			l, err = strconv.ParseInt(info[1], 0, 64)
			if err != nil {
				return nil, fmt.Errorf("failed parsing integer: %s", err)
			}
		}
		if len(info) > 2 && info[2] != "" {
			f, err = strconv.ParseInt(info[2], 0, 64)
			if err != nil {
				return nil, fmt.Errorf("failed parsing integer: %s", err)
			}
		}
		if len(info) > 3 && info[3] != "" {
			j = info[3]
		}

		instructionSrcMap[index] = &InstructionMapping{
			Start:     int(s),
			Length:    int(l),
			FileIndex: int(f),
			Jump:      j,
		}
	}

	for _, instruction := range instructionSrcMap {
		if instruction == nil {
			// Instruction does not map to source
			continue
		}

		i := 0
		line := 1
		column := 1

		for i < instruction.Start {
			if source[i] == '\n' {
				line++
				column = 0
			}

			column++
			i++
		}

		instruction.Line = line
		instruction.Column = column
	}

	memSrcMap, err := convertToMemoryMap(instructionSrcMap, bytecode)
	if err != nil {
		return nil, fmt.Errorf("failed mapping source map to memory: %s", err)
	}

	return &memSrcMap, nil
}

func convertToMemoryMap(sourceMap SourceMap, binData string) (SourceMap, error) {
	memSrcMap := make(SourceMap)

	if strings.HasPrefix(binData, "0x") {
		binData = binData[2:]
	}

	bin, err := hex.DecodeString(binData)
	if err != nil {
		return nil, fmt.Errorf("failed decoding runtime binary: %s", err)
	}

	instruction := 0
	for i := 0; i < len(bin); i++ {

		op := vm.OpCode(bin[i])
		extraPush := 0
		if op.IsPush() {
			// Skip more here
			extraPush = int(op - vm.PUSH1 + 1)
		}

		memSrcMap[i] = sourceMap[instruction]

		instruction++
		i += extraPush
	}

	return memSrcMap, nil
}
