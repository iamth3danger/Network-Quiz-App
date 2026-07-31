package services

import (
	"cmp"
	"errors"
	"math"
	"math/rand/v2"
	"slices"
)

type Generator string

const TotalQuestions = 10

var ErrInvalidGenerator = errors.New("Invalid generator in auto gen")

const (
	GeneratorNetwork       Generator = "network"
	GeneratorBroadcast     Generator = "broadcast"
	GeneratorFirstLastHost Generator = "firstlasthost"
	GeneratorHost          Generator = "host"
	GeneratorSubnet        Generator = "subnet"
	GeneratorCidr          Generator = "cidr"
)

var Generators = []Generator{GeneratorNetwork, GeneratorBroadcast, GeneratorFirstLastHost, GeneratorHost, GeneratorSubnet, GeneratorCidr}

type CidrGenerator struct {
	scoreMap map[Generator]Scorer
}

func autoGenerateQuestion(gen Generator) (Question, error) {
	switch gen {
	case GeneratorNetwork:
		return IPCidrToNetworkQuestion()
	case GeneratorBroadcast:
		return IPCidrToBroadcastQuestion()
	case GeneratorFirstLastHost:
		return IPCidrToFirstLastHostQuestion()
	case GeneratorHost:
		return CidrToHostQuestion()
	case GeneratorSubnet:
		return CidrToSubnetQuestion()
	case GeneratorCidr:
		return SubnetToCidrQuestion(), nil
	default:
		return Question{}, ErrInvalidGenerator
	}
}

type GenStat struct {
	generator Generator
	need      float64
}

func (c *CidrGenerator) getQuestionNum() map[Generator]int {
	genQuestions := make(map[Generator]int)
	var stats []GenStat
	needSum := 0.0
	count := 0

	for k, v := range c.scoreMap {
		val := float64(v.getRight()-v.getWrong()) + (float64(v.getStreak()) / float64(max((v.getRight()+v.getWrong())*10, 1)))

		need := (1 / math.Pow((val+0.1), 2))
		needSum += need
		stats = append(stats, GenStat{
			generator: k,
			need:      need,
		})
	}

	slices.SortFunc(stats, func(a, b GenStat) int {
		return cmp.Compare(a.need, b.need)
	})

	for _, s := range stats {
		questionNum := int(math.Floor((10 * s.need) / needSum))
		if questionNum == 0 && count < TotalQuestions {
			questionNum++
		}
		count++
		genQuestions[s.generator] = questionNum
	}

	for i := range TotalQuestions - count {
		genQuestions[stats[i].generator] = genQuestions[stats[i].generator] + 1
	}

	return genQuestions
}

func (c *CidrGenerator) createQuiz() ([]Question, error) {
	genQuestions := c.getQuestionNum()
	var quiz []Question

	for k, v := range genQuestions {
		for range v {
			question, err := autoGenerateQuestion(k)
			if err != nil {
				return nil, err
			}

			quiz = append(quiz, question)
		}
	}

	rand.Shuffle(len(quiz), func(i, j int) {
		quiz[i], quiz[j] = quiz[j], quiz[i]
	})

	return quiz, nil
}
