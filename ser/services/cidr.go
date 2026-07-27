package services

import (
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"
)

const SubnetMasks = 31
const Oct = 8
const SubnetLength = 4
const Subnets = 3
const OctectNum = 3
const QuestionNum = 4
const QuestionIter = 3

type DottedDecimal uint32

func (d DottedDecimal) string() (string, error) {
	networkOcts := make([]string, 4)

	for i := 3; i >= 0; i-- {
		octect := (d >> (8 * uint32(i))) & 0xFF
		networkOcts[3-i] = strconv.Itoa(int(octect))
	}

	return strings.Join(networkOcts, "."), nil
}

type IPConv func(uint32, uint32) uint32

func getAnswer(answer int32) string {
	switch answer {
	case 0:
		return "a"
	case 1:
		return "b"
	case 2:
		return "c"
	case 3:
		return "d"
	default:
		return "a"
	}
}

func CidrToSubnet(cidr int32) string {
	mask := uint32(0xFFFFFFFF) << (32 - uint32(cidr))
	octects := make([]string, 4)

	for i := 3; i >= 0; i-- {
		octect := (mask >> (8 * uint32(i))) & 0xFF
		octects[3-i] = strconv.Itoa(int(octect))
	}

	return strings.Join(octects, ".")
}

func octectsByte(octs []string) ([]byte, error) {
	var octects []byte
	for _, oct := range octs {
		octNum, err := strconv.ParseInt(oct, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parsing oct: %v", err)
		}
		octects = append(octects, byte(octNum))
	}

	return octects, nil
}

func addressToBits(addr string) (uint32, error) {
	octStr := strings.Split(addr, ".")
	octects, err := octectsByte(octStr)
	if err != nil {
		return 0, fmt.Errorf("%v", err)
	}

	return (uint32(octects[0]) << 24) | (uint32(octects[1]) << 16) | (uint32(octects[2]) << 8) | (uint32(octects[3])), nil
}

func IPCidrToNetwork(addr string, cidr int) (string, error) {
	return calcIPCidr(addr, cidr, func(ip, mask uint32) uint32 {
		return ip & mask
	})
}

func IPCidrToBroadcast(addr string, cidr int) (string, error) {
	return calcIPCidr(addr, cidr, func(ip, mask uint32) uint32 {
		return ip | ^mask
	})
}

func calcIPCidr(addr string, cidr int, conv IPConv) (string, error) {
	mask := uint32(0xFFFFFFFF) << (32 - uint32(cidr))
	ip, err := addressToBits(addr)

	if err != nil {
		return "", err
	}

	result := conv(ip, mask)

	return DottedDecimal(result).string()
}

func generateCidrs() []int32 {
	cidrSet := rand.Int32N(OctectNum)

	var cidrs []int32

	for range QuestionNum {
		cidrs = append(cidrs, cidrSet+rand.Int32N(Oct))
	}

	return cidrs
}

type valueGen func(cidr int32) (string, error)
type questionText func(cidr int32) string

func generateQuestion(gen valueGen, text questionText) (Question, error) {
	cidrs := generateCidrs()
	answer := rand.Int32N(QuestionIter)

	var results []string

	for i := range QuestionNum {
		res, err := gen(cidrs[i])

		if err != nil {
			return Question{}, err
		}

		results = append(results, res)
	}

	return Question{
		Question: text(cidrs[answer]),
		A:        results[0],
		B:        results[1],
		C:        results[2],
		D:        results[3],
		Answer:   getAnswer(answer),
	}, nil
}

func ToNetworkQuestion() (Question, error) {
	ip := DottedDecimal(rand.Uint32())
	ipStr, err := ip.string()

	if err != nil {
		return Question{}, err
	}

	return generateQuestion(
		func(cidr int32) (string, error) {
			return IPCidrToNetwork(ipStr, int(cidr))
		},
		func(cidr int32) string {
			return fmt.Sprintf("what is %s/%d as a network address?", ipStr, cidr)
		},
	)
}

func ToBroadcastQuestion() (Question, error) {
	ip := DottedDecimal(rand.Uint32())
	ipStr, err := ip.string()

	if err != nil {
		return Question{}, err
	}

	return generateQuestion(
		func(cidr int32) (string, error) {
			return IPCidrToBroadcast(ipStr, int(cidr))
		},
		func(cidr int32) string {
			return fmt.Sprintf("what is %s/%d as a broadcast address?", ipStr, cidr)
		},
	)
}

func CidrToSubnetQuestion() (Question, error) {

	return generateQuestion(
		func(cidr int32) (string, error) {
			return CidrToSubnet(cidr), nil
		},
		func(cidr int32) string {
			return fmt.Sprintf("What is /%d in dotted decimal?", cidr)
		},
	)
}

func SubnetToCidrQuestion() Question {
	cidrSet := rand.Int32N(OctectNum)
	answer := rand.Int32N(QuestionIter)

	var cidrs []int32

	for range QuestionNum {
		cidrs = append(cidrs, cidrSet+rand.Int32N(Oct))
	}

	answerSubnet := CidrToSubnet(cidrs[answer])
	question := fmt.Sprintf("Convert %s to CIDR", answerSubnet)

	cidrStr := func(cidr int32) string {
		return fmt.Sprintf("/%d", cidr)
	}

	return Question{
		Question: question,
		A:        cidrStr(cidrs[0]),
		B:        cidrStr(cidrs[1]),
		C:        cidrStr(cidrs[2]),
		D:        cidrStr(cidrs[3]),
		Answer:   getAnswer(answer),
	}
}
