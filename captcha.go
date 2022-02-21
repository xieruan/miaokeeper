package main

import (
	"bytes"
	"math/rand"

	"github.com/steambap/captcha"
)

type CaptchaType uint

const (
	CaptchaTypeNormal CaptchaType = iota
	CaptchaTypeMathExpr
)

func GenerateRandomCaptcha() (*bytes.Buffer, []string) {
	if rand.Intn(10) >= 5 {
		buf, result := GenerateCaptcha(CaptchaTypeMathExpr)
		return buf, GenerateCaptchaOptions(CaptchaTypeMathExpr, result)
	}
	buf, result := GenerateCaptcha(CaptchaTypeNormal)
	return buf, GenerateCaptchaOptions(CaptchaTypeNormal, result)
}

func GenerateCaptcha(captchaType CaptchaType) (*bytes.Buffer, string) {
	var data *captcha.Data = nil

	if captchaType == CaptchaTypeNormal {
		data, _ = captcha.New(330, 110)
	} else {
		data, _ = captcha.NewMathExpr(330, 110)
	}

	buf := bytes.Buffer{}
	data.WriteImage(&buf)
	return &buf, data.Text
}

var ObfsExcludes = [][]string{
	{"O", "0", "Q", "D", "o"},
	{"1", "I", "l"},
	{"2", "z", "Z"},
}
var ObfsSet = []rune("1234567890qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM")

func GenerateCaptchaRandomize() rune {
	return ObfsSet[rand.Intn(len(ObfsSet))]
}

func GenerateCaptchaReplacableChar(s rune) rune {
	excludes := []string{}
	for _, ex := range ObfsExcludes {
		if ContainsString(ex, string(s)) {
			excludes = ex
		}
	}

	attempts := 0
	for {
		attempts += 1
		s := GenerateCaptchaRandomize()
		if !ContainsString(excludes, string(s)) {
			return s
		}
		if attempts > 10 {
			return 'k'
		}
	}
}

func GenerateCaptchaOptions(captchaType CaptchaType, original string) []string {
	waitList := []string{original}

	for i := 0; i < 3; i++ {
		attempts := 2
		target := []rune(original)
		if captchaType == CaptchaTypeMathExpr {
			attempts = 1
		}
		for j := 0; j < attempts; j += 1 {
			index := rand.Intn(len(target))
			target[index] = GenerateCaptchaReplacableChar(target[index])
		}
		waitList = append(waitList, string(target))
	}

	return waitList
}
