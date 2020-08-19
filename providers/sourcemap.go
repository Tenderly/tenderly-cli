package providers

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/tenderly/tenderly-cli/stacktrace"
)

func ParseContract(contract *Contract) (stacktrace.SourceMap, error) {
	rawSrcMap := contract.DeployedSourceMap
	instructionSrcMap, err := parseInstructionSourceMap(rawSrcMap)
	if err != nil {
		return nil, fmt.Errorf("sourcemap.Parse: %s", err)
	}

	rawBytecode := contract.DeployedBytecode

	memSrcMap, err := convertToMemoryMap(instructionSrcMap, rawBytecode)
	if err != nil {
		return nil, fmt.Errorf("sourcemap.Parse: %s", err)
	}

	rawSrc := contract.Source

	for _, instruction := range memSrcMap {
		if instruction == nil {
			// Instruction does not map to source
			continue
		}

		i := 0
		line := 1
		column := 1

		for i < instruction.Start {
			if rawSrc[i] == '\n' {
				line++
				column = 0
			}

			column++
			i++

			if i == len(rawSrc)-1 {
				break
			}
		}

		instruction.Line = line
		instruction.Column = column
	}

	return memSrcMap, nil
}

func Parse(contracts map[string]*Contract) (map[string]stacktrace.SourceMap, map[string][]byte, error) {
	sourceMaps := make(map[string]stacktrace.SourceMap)
	binaries := make(map[string][]byte)
	for key, contract := range contracts {
		if contract != nil {
			rawSrcMap := contract.DeployedSourceMap
			instructionSrcMap, err := parseInstructionSourceMap(rawSrcMap)
			if err != nil {
				return nil, nil, fmt.Errorf("sourcemap.Parse: %s", err)
			}

			rawBytecode := contract.DeployedBytecode

			bin, err := hex.DecodeString(rawBytecode[2:])
			if err != nil {
				return nil, nil, fmt.Errorf("failed decoding runtime binary: %s", err)
			}

			memSrcMap, err := convertToMemoryMap(instructionSrcMap, rawBytecode)
			if err != nil {
				return nil, nil, fmt.Errorf("sourcemap.Parse: %s", err)
			}

			rawSrc := contract.Source

			for _, instruction := range memSrcMap {
				if instruction == nil {
					// Instruction does not map to source
					continue
				}

				i := 0
				line := 1
				column := 1

				for i < instruction.Start {
					if rawSrc[i] == '\n' {
						line++
						column = 0
					}

					column++
					i++
				}

				instruction.Line = line
				instruction.Column = column
			}

			sourceMaps[key] = memSrcMap
			binaries[key] = bin
		}
	}

	return sourceMaps, binaries, nil
}

func parseInstructionSourceMap(rawSrcMap string) (stacktrace.SourceMap, error) {
	instructionSrcMap := make(stacktrace.SourceMap)

	var err error
	var s int64
	var l int64
	var f int64
	var j string
	for index, mapping := range strings.Split(rawSrcMap, ";") {
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

		instructionSrcMap[index] = &stacktrace.InstructionMapping{
			Start:     int(s),
			Length:    int(l),
			FileIndex: int(f),
			Jump:      j,
		}
	}

	return instructionSrcMap, nil
}

func convertToMemoryMap(sourceMap stacktrace.SourceMap, binData string) (stacktrace.SourceMap, error) {
	memSrcMap := make(stacktrace.SourceMap)

	bin, err := hex.DecodeString(binData[2:])
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
