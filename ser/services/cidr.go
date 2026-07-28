package services

import (
	"fmt"
	"math"
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

type Choices struct {
	results []int32
	answer  int32
}

func genChoices() Choices {
	return Choices{
		results: genCidersFromRandomRange(),
		answer:  rand.Int32N(QuestionIter),
	}
}

func genHostChoices() Choices {
	return Choices{
		results: generateCidrs(3),
		answer:  rand.Int32N(QuestionIter),
	}
}

type ConvChoices struct {
	cidrs     []int32
	ip        uint32
	answer    int32
	addresses []string
}

func (c ConvChoices) getAnswerCider() int32 {
	return c.cidrs[c.answer]
}

func getConvChoices(conv func(uint32, int) uint32) (ConvChoices, error) {
	ip := rand.Uint32()
	cidrSet := rand.Int32N(OctectNum)
	answer := rand.Int32N(QuestionIter)

	var cidrs []int32
	var addresses []string
	seenCidrs := make(map[int32]struct{})
	seenAddr := make(map[string]struct{})

	for len(cidrs) < 4 {

		if len(seenCidrs) > 0 && len(seenCidrs)%8 == 0 {
			cidrSet--
			if cidrSet < 0 {
				cidrSet = 3
			}
		}

		cidr := (Oct * cidrSet) + rand.Int32N(Oct)

		_, ok := seenCidrs[cidr]
		if ok {
			continue
		}
		seenCidrs[cidr] = struct{}{}
		addr := conv(ip, int(cidr))
		address, err := DottedDecimal(addr).string()
		if err != nil {
			return ConvChoices{}, err
		}
		_, ok = seenAddr[address]
		if ok {
			continue
		}

		seenAddr[address] = struct{}{}
		cidrs = append(cidrs, cidr)
		addresses = append(addresses, address)
	}

	return ConvChoices{
		cidrs:     cidrs,
		ip:        ip,
		answer:    answer,
		addresses: addresses,
	}, nil
}

type valueConversion func(cidr int32) (string, error)
type generator func() Choices
type questionText func(cidr int32) string

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

func cidrToHosts(cidr int32) string {
	return strconv.Itoa(int(math.Pow(2, float64((32-cidr)))) - 2)
}

func IPCidrToNetwork(ip uint32, cidr int) uint32 {
	mask := uint32(0xFFFFFFFF) << (32 - uint32(cidr))
	return ip & mask
}

func IPCidrToBroadcast(ip uint32, cidr int) uint32 {
	mask := uint32(0xFFFFFFFF) << (32 - uint32(cidr))
	return ip | ^mask
}

func genCidersFromRandomRange() []int32 {
	octNum := rand.Int32N(OctectNum)
	return generateCidrs(octNum)
}

func generateCidrs(octNum int32) []int32 {

	var cidrs []int32
	seen := make(map[int32]struct{})

	for len(cidrs) < 4 {
		cidr := (Oct * octNum) + rand.Int32N(Oct)

		if cidr == 0 {
			cidr++
		}
		_, ok := seen[cidr]
		if ok {
			continue
		}
		seen[cidr] = struct{}{}
		cidrs = append(cidrs, cidr)
	}

	return cidrs
}

type iPCidrConverter func(uint32, int) uint32

func iPCidrConversionQuestion(conv iPCidrConverter, addrType string) (Question, error) {
	ch, err := getConvChoices(conv)
	if err != nil {
		return Question{}, err
	}

	ipStr, err := DottedDecimal(ch.ip).string()

	return Question{
		Question: fmt.Sprintf("what is %s/%d as a %s address?", ipStr, ch.getAnswerCider(), addrType),
		A:        ch.addresses[0],
		B:        ch.addresses[1],
		C:        ch.addresses[2],
		D:        ch.addresses[3],
		Answer:   getAnswer(ch.answer),
	}, nil
}

func IPCidrToNetworkQuestion() (Question, error) {
	return iPCidrConversionQuestion(IPCidrToNetwork, "network")
}

func IPCidrToBroadcastQuestion() (Question, error) {
	return iPCidrConversionQuestion(IPCidrToBroadcast, "broadcast")
}

func generateQuestion(conv valueConversion, text questionText, ch Choices) (Question, error) {
	cidrs := ch.results
	answer := ch.answer

	var results []string

	for i := range QuestionNum {
		res, err := conv(cidrs[i])

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

func CidrToHostQuestion() (Question, error) {
	return generateQuestion(
		func(cidr int32) (string, error) {
			return cidrToHosts(cidr), nil
		},
		func(cidr int32) string {
			return fmt.Sprintf("How many usable hosts does /%d have?", cidr)
		},
		genHostChoices(),
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
		genChoices(),
	)
}

func SubnetToCidrQuestion() Question {
	cidrs := genCidersFromRandomRange()
	answer := rand.Int32N(QuestionIter)

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
