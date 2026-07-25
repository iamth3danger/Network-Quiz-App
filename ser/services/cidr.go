package services

import (
	"fmt"
	"math/rand"
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

func CidrToSubnetQuestion() Question {
	cidrSet := rand.Int31n(OctectNum)
	answer := rand.Int31n(QuestionIter)

	var cidrs []int32
	var subnets []string

	for range QuestionNum {
		cidrs = append(cidrs, cidrSet+rand.Int31n(Oct))
	}

	for i := range QuestionNum {
		subnets = append(subnets, CidrToSubnet(cidrs[i]))
	}

	question := fmt.Sprintf("What is /%d in dotted decimal?", int(cidrs[answer]))

	return Question{
		Question: question,
		A:        subnets[0],
		B:        subnets[1],
		C:        subnets[2],
		D:        subnets[3],
		Answer:   getAnswer(answer),
	}
}

func SubnetToCidrQuestion() Question {
	cidrSet := rand.Int31n(OctectNum)
	answer := rand.Int31n(QuestionIter)

	var cidrs []int32

	for range QuestionNum {
		cidrs = append(cidrs, cidrSet+rand.Int31n(Oct))
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
