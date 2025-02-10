package partialjson

/*
 * Copyright (c) 2025 shado1111w.
 * Licensed under the MIT License.
 * See LICENSE file in the project root for full license information.
 */

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

const testData = `{"roles":[{"role_name":"我","role_desc":"女，青年，DJ，坚毅"},{"role_name":"墨镜僵尸","role_desc":"男，青年，僵尸团队成员，富有经验"},{"role_name":"领舞尸王","role_desc":"男，青年，僵尸团队领舞者，敬业"},{"role_name":"小僵尸","role_desc":"男，少年，僵尸团队成员，天真"},{"role_name":"飘逸之神霹雳飞天腿","role_desc":"男，青年，舞蹈界传奇，狂热"}],"scene_list":[{"screen_description":"夜晚，小区广场灯火微明，僵尸们的练习场一片热闹，荧光服闪烁，气氛紧张刺激。","chat_group":[{"role_name":"我","content":"我盯着墨镜僵尸那张忧愁的脸，忍不住问：“下一场是谁来的？还能比步王郎更炸场？”","emotion":"疑惑"},{"role_name":"墨镜僵尸","content":"“对方是舞蹈界的传说——‘飘逸之神’霹雳飞天腿！据说他能把舞步跳出粒子分解效果，一甩腿，整个广场都能变成荧光沙滩！”","emotion":"担忧"},{"role_name":"我","content":"闻言，我倒吸一口凉气：“那我们岂不是要凉了？”但随即一拍脑门：“不行！咱们僵尸乐队不能轻易认输！既然这次对手如此强大，那就得想出一招绝杀！”","emotion":"坚定"},{"role_name":"旁白","content":"僵尸们齐齐凑过来围成一圈，纷纷踊跃出谋划策。领舞尸王率先发言。","emotion":"期待"},{"role_name":"领舞尸王","content":"“要不我们练习‘乾坤大挪移连环甩头法’，一甩甩出宇宙轨迹？”","emotion":"提议"},{"role_name":"墨镜僵尸","content":"摸着僵尸巴：“不不不，咱们得突出创意！比如全场倒立跳舞，上下颠倒也能横扫全场！”","emotion":"沉思"},{"role_name":"我","content":"“不够疯狂！对方能跳出粒子效果，我们得更炫、更炸、更令人拍烂手掌！”","emotion":"焦急"},{"role_name":"小僵尸","content":"怯生生举手：“那个……我觉得可以加点情怀，比如，跳一支人人都会但没人想到的怀旧舞？”","emotion":"犹豫"},{"role_name":"旁白","content":"小僵尸的提议点醒了我！怀旧与创新结合，这不就是绝杀吗！我激动地站起来拍桌子。","emotion":"兴奋"},{"role_name":"我","content":"“就这么定了！咱们去复刻20世纪最经典的街舞《千手观音》，然后配上炫酷特效，把全场炸成大佛光环！”","emotion":"决心"}]}],"question":"如何面对外星舞王的挑战？","options":["接受挑战，与外星人切磋舞技","拒绝挑战，专注地球舞台发展"]}`

var jsonTestDataList = make([]string, len([]rune(testData)))

func init() {
	c := ""
	for i, data := range []rune(testData) {
		c += string(data)
		jsonTestDataList[i] = c
	}
}

func TestParseSpace(t *testing.T) {
	parser := NewJSONParser(true)

	tests := []struct {
		input, remaining string
	}{
		{
			input:     " ",
			remaining: "",
		},
		{
			input:     "  ",
			remaining: "",
		},
		{
			input:     "\r",
			remaining: "",
		},
		{
			input:     "\n",
			remaining: "",
		},
		{
			input:     "\t",
			remaining: "",
		},
		{
			input:     "  \r\n\t",
			remaining: "",
		},
	}

	for _, test := range tests {
		_, remaining, err := parser.parseSpace(test.input)
		require.Nil(t, err)

		require.Equal(t, test.remaining, remaining)
	}
}

func TestParseArray(t *testing.T) {
	parser := NewJSONParser(true)

	tests := []struct {
		input    string
		expected interface{}
		err      error
	}{
		{
			input:    "[",
			expected: nil,
		},
		{
			input: "[1, abc",
			err:   ErrUnexpectedToken,
		},
		{
			input:    "[1, 2, \"3\"",
			expected: []interface{}{1, 2, "3"},
		},
		{
			input:    "[1, 2, 3",
			expected: []interface{}{1, 2, 3},
		},
		{
			input:    "[\"1\", \"2\", \"3\"",
			expected: []interface{}{"1", "2", "3"},
		},
	}

	for _, test := range tests {
		obj, _, err := parser.parseArray(test.input)
		require.Equal(t, test.err, err)

		if err == nil {
			expected, ok1 := test.expected.([]interface{})
			actual, ok2 := obj.([]interface{})
			if ok1 && ok2 {
				equalArray(t, expected, actual)
			} else {
				require.EqualValues(t, test.expected, obj)
			}
		}
	}
}

func equalArray(t *testing.T, expected, actual []interface{}) {
	require.Equal(t, len(expected), len(actual))

	for i := range expected {
		require.EqualValues(t, expected[i], actual[i])
	}
}

func TestParseString(t *testing.T) {
	tests := []struct {
		input, expected string
		strict          bool
		err             error
	}{
		{
			input:  "\"你好",
			err:    ErrIncompleteString,
			strict: true,
		},
		{
			input:    "\"你好\"",
			expected: "你好",
			strict:   true,
		},
		{

			input:    "\"你好，\\\"世界\\\"。\"",
			expected: "你好，\"世界\"。",
			strict:   true,
		},
		{
			input:    "\"你好",
			expected: "你好",
			strict:   false,
		},
		{
			input:    "\"你好，\\\"世界\\\"。",
			expected: "你好，\\\"世界\\\"。",
			strict:   false,
		},
	}

	for _, test := range tests {
		parser := NewJSONParser(test.strict)
		obj, _, err := parser.parseString(test.input)
		require.Equal(t, test.err, err)

		if err == nil {
			require.EqualValues(t, test.expected, obj)
		}
	}
}

func TestParseNum(t *testing.T) {
	parser := NewJSONParser(true)

	tests := []struct {
		input    string
		expected float64
		err      error
	}{
		{
			input:    "1.23e-4",
			expected: 0.000123,
		},
		{
			input:    "1.23e+4",
			expected: 12300,
		},
		{
			input:    "-12.34",
			expected: -12.34,
		},
		{
			input:    "12.34",
			expected: 12.34,
		},
		{
			input: "12.1e",
			err:   ErrIncompleteNum,
		},
		{
			input: "-",
			err:   ErrIncompleteNum,
		},
	}

	for _, tc := range tests {
		obj, _, err := parser.parseNumber(tc.input)
		require.Equal(t, tc.err, err)

		if err == nil {
			require.EqualValues(t, tc.expected, obj)
		}
	}
}

func TestParseTrue(t *testing.T) {
	parser := NewJSONParser(true)

	tests := []struct {
		input    string
		expected bool
		err      error
	}{
		{
			input:    "true",
			expected: true,
		},
		{
			input: "tru",
			err:   ErrUnexpectedToken,
		},
	}

	for _, tc := range tests {
		obj, _, err := parser.parseTrue(tc.input)
		require.Equal(t, tc.err, err)

		if err == nil {
			require.EqualValues(t, tc.expected, obj)
		}
	}
}

func TestParseFalse(t *testing.T) {
	parser := NewJSONParser(true)

	tests := []struct {
		input    string
		expected bool
		err      error
	}{
		{
			input:    "false",
			expected: false,
		},
		{
			input: "fals",
			err:   ErrUnexpectedToken,
		},
	}

	for _, tc := range tests {
		obj, _, err := parser.parseFalse(tc.input)
		require.Equal(t, tc.err, err)

		if err == nil {
			require.EqualValues(t, tc.expected, obj)
		}
	}
}

func TestParseNull(t *testing.T) {
	parser := NewJSONParser(true)
	tests := []struct {
		input    string
		expected any
		err      error
	}{
		{
			input:    "null",
			expected: nil,
		},
		{
			input: "nul",
			err:   ErrUnexpectedToken,
		},
	}

	for _, test := range tests {
		obj, _, err := parser.parseNull(test.input)
		require.Equal(t, test.err, err)

		if err == nil {
			require.EqualValues(t, test.expected, obj)
		}
	}
}

func TestEnsureJson(t *testing.T) {
	tests := []struct {
		input, expected string
		strict          bool
		err             error
	}{
		{
			input:    "{\"nam",
			expected: "{}",
			strict:   false,
		},
		{
			input:    "{\"你好，\\\"世界\\\"。",
			expected: "{}",
			strict:   false,
		},
		{
			input:    "{\"name\":\"Alice",
			expected: "{\"name\":\"Alice\"}",
			strict:   false,
		},
		{
			input:  "",
			err:    ErrUnexpectedToken,
			strict: true,
		},
		{
			input:  "1",
			err:    ErrUnexpectedToken,
			strict: true,
		},
		{
			input:  `{"options":["\"是我自己清晰的脸\"", abc`,
			err:    ErrUnexpectedToken,
			strict: true,
		},
		{
			input:  `{"options":["\"是我自己清晰的脸\""], 123`,
			err:    ErrUnexpectedToken,
			strict: true,
		},
		{
			input:  `{"options",["\"是我自己清晰的脸\""], 123`,
			err:    ErrUnexpectedToken,
			strict: true,
		},
		{
			input:    `["是我自己清晰的脸","是初中`,
			expected: `["是我自己清晰的脸"]`,
			strict:   true,
		},
		{
			input:    "{",
			expected: "{}",
			strict:   true,
		},
		{
			input:    "{\"nam",
			expected: "{}",
			strict:   true,
		},
		{
			input:    `{"options":"\"是我自己清晰的脸\`,
			expected: `{"options":null}`,
			strict:   true,
		},
		{
			input:    "{\"options\":[\"是已故奶奶的脸\",\"是过去自己的脸\"],\"question\":\"你看到[{}]}}熟悉的脸是谁的\",\"roles\":[{\"",
			expected: "{\"options\":[\"是已故奶奶的脸\",\"是过去自己的脸\"],\"question\":\"你看到[{}]}}熟悉的脸是谁的\",\"roles\":null}",
			strict:   true,
		},
		{
			input:    `{"options":["\"是我自己清晰的脸\""`,
			expected: `{"options":["\"是我自己清晰的脸\""]}`,
			strict:   true,
		},
	}

	for _, test := range tests {
		parser := NewJSONParser(test.strict)
		data, err := parser.EnsureJSON(test.input)
		require.Equal(t, test.err, err, test.input+test.expected)

		if err == nil {
			require.Equal(t, test.expected, data)
		}
	}
}

func TestUnmarshal(t *testing.T) {
	parser := NewJSONParser(true, WithOnExtraToken(func(text string, data any, remaining string) {
		fmt.Printf("Parsed JSON with extra tokens: text: %s, data: %v, reminding: %s\n", text, data, remaining)
	}))
	tests := []struct {
		input    string
		expected *testObject
		err      error
	}{
		{
			input:    "{",
			expected: &testObject{},
		},
		{
			input:    `{"options":"\"是我自己清晰的脸\`,
			expected: &testObject{},
		},
		{
			input:    "{\"options\":[\"是已故奶奶的脸\",\"是过去自己的脸\"],\"question\":\"你看到[{}]}}熟悉的脸是谁的\",\"roles\":[{\"",
			expected: &testObject{Options: []string{"是已故奶奶的脸", "是过去自己的脸"}, Question: "你看到[{}]}}熟悉的脸是谁的"},
		},
		{
			input:    `{"options":["\"是我自己清晰的脸\""`,
			expected: &testObject{Options: []string{"\"是我自己清晰的脸\""}},
		},
	}

	for _, test := range tests {
		obj := testObject{}
		err := parser.Unmarshal([]byte(test.input), &obj)
		require.Equal(t, test.err, err)

		if err == nil {
			require.EqualValues(t, test.expected, &obj)
		}
	}
}

func TestFastUnmarshal(t *testing.T) {
	parser := NewJSONParser(true, WithOnExtraToken(func(text string, data any, remaining string) {
		fmt.Printf("Parsed JSON with extra tokens: text: %s, data: %v, reminding: %s\n", text, data, remaining)
	}))
	tests := []struct {
		input    string
		expected *testObject
		err      error
	}{
		{
			input:    "{",
			expected: &testObject{},
		},
		{
			input:    `{"options":"\"是我自己清晰的脸\`,
			expected: &testObject{},
		},
		{
			input:    "{\"options\":[\"是已故奶奶的脸\",\"是过去自己的脸\"],\"question\":\"你看到[{}]}}熟悉的脸是谁的\",\"roles\":[{\"",
			expected: &testObject{Options: []string{"是已故奶奶的脸", "是过去自己的脸"}, Question: "你看到[{}]}}熟悉的脸是谁的"},
		},
		{
			input:    `{"options":["\"是我自己清晰的脸\""`,
			expected: &testObject{Options: []string{"\"是我自己清晰的脸\""}},
		},
	}

	for _, test := range tests {
		obj := testObject{}
		err := parser.FastUnmarshal([]byte(test.input), &obj)
		require.Equal(t, test.err, err)

		if err == nil {
			require.EqualValues(t, test.expected, &obj)
		}
	}
}

type testObject struct {
	Options  []string `json:"options"`
	Question string   `json:"question"`
	Roles    []struct {
		RoleDesc string `json:"role_desc"`
		RoleName string `json:"role_name"`
	} `json:"roles"`
	SceneList []struct {
		ChatGroup []struct {
			Content  string `json:"content"`
			Emotion  string `json:"emotion"`
			RoleName string `json:"role_name"`
		} `json:"chat_group"`
		ScreenDescription string `json:"screen_description"`
	} `json:"scene_list"`
}

func TestFastEnsureJsonResultValid(t *testing.T) {
	parser := NewJSONParser(true, WithDefaultOnExtraToken())
	for _, testData := range jsonTestDataList {
		data, err := parser.EnsureJSON(testData)
		require.Nil(t, err)

		obj := testObject{}
		err = json.Unmarshal([]byte(data), &obj)
		require.Nil(t, err)

		fastData, err := parser.FastEnsureJSON(testData)
		require.Nil(t, err, testData)

		fastObj := testObject{}
		err = json.Unmarshal([]byte(fastData), &fastObj)
		require.Nil(t, err)

		require.EqualValues(t, obj, fastObj)
	}
}

func BenchmarkEnsureJson(b *testing.B) {
	parser := NewJSONParser(true)
	for i := 0; i < b.N; i++ {
		for _, testData := range jsonTestDataList {
			_, err := parser.EnsureJSON(testData)
			require.Nil(b, err)
		}
	}
}

func BenchmarkFastEnsureJson(b *testing.B) {
	parser := NewJSONParser(true)
	for i := 0; i < b.N; i++ {
		for _, testData := range jsonTestDataList {
			_, err := parser.FastEnsureJSON(testData)
			require.Nil(b, err)
		}
	}
}
