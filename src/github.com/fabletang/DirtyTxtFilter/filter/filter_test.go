package filter

import (
	"io"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

var filter2 *Filter

func init() {
	filter2 := New()
	err := filter2.LoadWordDict(".././res/dict/dict.txt")
	if err != nil {
		println("fail to load dict %v", err)
	} else {
		println("init filter...")
	}
}

func TestLoadDict(t *testing.T) {
	filter := New()
	err := filter.LoadWordDict(".././res/dict/dict.txt")
	if err != nil {
		t.Errorf("fail to load dict %v", err)
	}
}

func TestLoadNetWordDict(t *testing.T) {
	filter := New()
	err := filter.LoadNetWordDict("https://raw.githubusercontent.com/importcjj/sensitive/master/dict/dict.txt")
	if err != nil {
		//t.Errorf("fail to load dict %v", err)
	}
	if len(filter.trie.Root.Children) == 0 {
		//t.Errorf("load dict empty")
	}
}

func TestLoad(t *testing.T) {
	filter := New()
	var r io.Reader
	r = strings.NewReader("read")
	err := filter.Load(r)
	if err != nil {
		t.Errorf("fail to load dict %v", err)
	}
	if len(filter.trie.Root.Children) == 0 {
		t.Errorf("load dict empty")
	}
}

func TestSensitiveFilter(t *testing.T) {
	filter := New()
	filter.AddWord("有一个东西")
	filter.AddWord("一个东西")
	filter.AddWord("一个")
	filter.AddWord("东西")
	filter.AddWord("个东")

	testcases := []struct {
		Text   string
		Expect string
	}{
		{"我有一个东东西", "我有东"},
		{"我有一个东西", "我"},
		{"一个东西", ""},
		{"两个东西", "两西"},
		{"一个物体", "物体"},
		//{"一*个物体", "物体"},
	}

	for _, tc := range testcases {
		if got := filter.Filter(tc.Text); got != tc.Expect {
			t.Fatalf("filter %s, got %s, expect %s", tc.Text, got, tc.Expect)
		}
	}

}
func TestCheckAndReplace(t *testing.T) {
	filter := New()
	filter.LoadWordDict(".././res/dict/dict.txt")
	//filter.AddWord("傻")
	//filter.AddWord("煞叉")

	testcases := []struct {
		Text       string
		ExpectPass bool
		Expect     string
	}{
		//{"😄傻",false, "傻"},
		{"f u c k", false, "****"},
		{"f u`c@k", false, "****"},
		{"😄煞🦌叉", false, "**"},
		{"😄煞(叉)", false, "**"},
		{"😄煞（叉）", false, "**"},
		{"😄   煞    (叉)", false, "**"},
		{"😄   煞    (b)", false, "**"},
		//{"🔪你",false, "*你"},
	}

	for _, tc := range testcases {
		if pass, first := filter.CheckAndReplace(tc.Text, '*'); pass != tc.ExpectPass || first != tc.Expect {
			t.Fatalf("check and replace:%s, got %v, %s, expect %v, %s", tc.Text, pass, first, tc.ExpectPass, tc.Expect)
		}
	}

}
func TestSensitiveValidateSingleword(t *testing.T) {
	filter := New()
	filter.AddWord("东")

	testcases := []struct {
		Text        string
		ExpectPass  bool
		ExpectFirst string
	}{
		{"两个东西", false, "东"},
	}

	for _, tc := range testcases {
		if pass, first := filter.Validate(tc.Text); pass != tc.ExpectPass || first != tc.ExpectFirst {
			t.Fatalf("validate %s, got %v, %s, expect %v, %s", tc.Text, pass, first, tc.ExpectPass, tc.ExpectFirst)
		}
	}

}

func TestSensitiveValidate(t *testing.T) {
	filter := New()
	//filter =filter2
	filter.AddWord("有一个东西")
	filter.AddWord("一个东西")
	filter.AddWord("一个")
	filter.AddWord("东西")
	filter.AddWord("个东")
	filter.AddWord("有一个东西")
	filter.AddWord("一个东西")
	filter.AddWord("一个")
	filter.AddWord("东西")
	filter.AddWord("fuck")
	filter.AddWord("FUCK")
	filter.AddWord("Fuck")
	filter.AddWord("法克油")
	filter.AddWord("法油")

	testcases := []struct {
		Text        string
		ExpectPass  bool
		ExpectFirst string
	}{
		{"我有一@ |个东东西", false, "一个"},
		{"我有一个东东西", false, "一个"},
		{"我有一个东西", false, "有一个东西"},
		{"一个东西", false, "一个"},
		{"两个东西", false, "个东"},
		{"一样东西", false, "东西"},
		//{"f*u*c|k", false, "fuck"},
		//{"F*u*c|k", false, "Fuck"},
		{"法*克 油", false, "法克油"},
		{"法🉑️油", false, "法油"},
		{"法【】油", false, "法油"},
		{"法😁油", false, "法油"},
		{"法（油）", false, "法油"},
		{"^_^", true, ""},
	}

	for _, tc := range testcases {
		if pass, first := filter.Validate(tc.Text); pass != tc.ExpectPass || first != tc.ExpectFirst {
			t.Fatalf("validate %s, got %v, %s, expect %v, %s", tc.Text, pass, first, tc.ExpectPass, tc.ExpectFirst)
		}
	}

}

//func TestFilterEmoji(t *testing.T) {
//	filter := New()
//	//filter =filter2
//	filter.AddWord("法油")
//	testcases := []struct {
//		Text        string
//		//ExpectPass  bool
//		ExpectFirst string
//	}{
//		//{"法🉑️油",  "法油"},
//		//{"法1油",  "法油"},
//		{"法【】油",  "法油"},
//		//{"法😁油",  "法油"},
//	}
//	for _, tc := range testcases {
//		if first := filter.FilterEmoji(tc.Text); first != tc.ExpectFirst {
//			t.Fatalf("validate %s, got %s, expect  %s", tc.Text,  first,  tc.ExpectFirst)
//		}
//	}
//
//}

func BenchmarkSensitiveValidate(b *testing.B) {
	//filter := New()
	filter := New()
	err := filter.LoadWordDict(".././dict/dict.txt")
	if err != nil {
		println("fail to load dict %v", err)
	}
	filter.AddWord("有一个东西")
	filter.AddWord("一个东西")
	filter.AddWord("一个")
	filter.AddWord("东西")
	filter.AddWord("个东")
	filter.AddWord("有一个东西")
	filter.AddWord("一个东西")
	filter.AddWord("一个")
	filter.AddWord("东西")
	filter.AddWord("fuck")
	filter.AddWord("FUCK")
	filter.AddWord("Fuck")
	filter.AddWord("法克油")
	filter.AddWord("法油")

	testcases := []struct {
		Text        string
		ExpectPass  bool
		ExpectFirst string
	}{
		{"我有一@ |个东东西", false, "一个"},
		{"我有一个东东西", false, "一个"},
		{"我有一个东西", false, "有一个东西"},
		{"一个东西", false, "一个"},
		{"两个东西", false, "个东"},
		{"一样东西", false, "东西"},
		//{"f*u*c|k", false, "fuck"},
		//{"F*u*c|k", false, "Fuck"},
		{"法*克 油", false, "法克油"},
		{"法🉑️油", false, "法油"},
		{"法【】油", false, "法油"},
		{"法😁油", false, "法油"},
	}

	for _, tc := range testcases {
		if pass, first := filter.Validate(tc.Text); pass != tc.ExpectPass || first != tc.ExpectFirst {
			b.Fatalf("validate %s, got %v, %s, expect %v, %s", tc.Text, pass, first, tc.ExpectPass, tc.ExpectFirst)
		}
	}

}

func TestSensitiveReplace(t *testing.T) {
	filter := New()
	filter.AddWord("有一个东西")
	filter.AddWord("一个东西")
	filter.AddWord("一个")
	filter.AddWord("东西")
	filter.AddWord("个东")

	testcases := []struct {
		Text   string
		Expect string
	}{
		{"我有一个东东西", "我有**东**"},
		{"我有一个东西", "我*****"},
		{"一个东西", "****"},
		{"两个东西", "两**西"},
		{"一个物体", "**物体"},
		//{"一*个物体", "**物体"},
	}

	for _, tc := range testcases {
		if got := filter.Replace(tc.Text, '*'); got != tc.Expect {
			t.Fatalf("replace %s, got %s, expect %s", tc.Text, got, tc.Expect)
		}
	}

}

func TestSensitiveFindAll(t *testing.T) {
	filter := New()
	filter.AddWord("有一个东西")
	filter.AddWord("一个东西")
	filter.AddWord("一个")
	filter.AddWord("东西")
	filter.AddWord("个东")

	testcases := []struct {
		Text   string
		Expect []string
	}{
		{"我有一个东东西", []string{"一个", "个东", "东西"}},
		{"我有一个东西", []string{"有一个东西", "一个", "一个东西", "个东", "东西"}},
		{"一个东西", []string{"一个", "一个东西", "个东", "东西"}},
		{"两个东西", []string{"个东", "东西"}},
		{"一个物体", []string{"一个"}},
	}

	for _, tc := range testcases {
		if got := filter.FindAll(tc.Text); !reflect.DeepEqual(tc.Expect, got) {
			t.Fatalf("findall %s, got %s, expect %s", tc.Text, got, tc.Expect)
		}
	}
}

func TestSensitiveFindallSingleword(t *testing.T) {
	filter := New()
	filter.AddWord("东")

	testcases := []struct {
		Text   string
		Expect []string
	}{
		{"两个东西", []string{"东"}},
	}

	for _, tc := range testcases {
		if got := filter.FindAll(tc.Text); !reflect.DeepEqual(tc.Expect, got) {
			t.Fatalf("findall %s, got %s, expect %s", tc.Text, got, tc.Expect)
		}
	}

}

func TestFilter_LoadWordDict(t *testing.T) {
	type fields struct {
		trie  *Trie
		noise *regexp.Regexp
	}
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := &Filter{
				trie:  tt.fields.trie,
				noise: tt.fields.noise,
			}
			if err := filter.LoadWordDict(tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("Filter.LoadWordDict() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFilter_LoadNetWordDict(t *testing.T) {
	type fields struct {
		trie  *Trie
		noise *regexp.Regexp
	}
	type args struct {
		url string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := &Filter{
				trie:  tt.fields.trie,
				noise: tt.fields.noise,
			}
			if err := filter.LoadNetWordDict(tt.args.url); (err != nil) != tt.wantErr {
				t.Errorf("Filter.LoadNetWordDict() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFilter_Load(t *testing.T) {
	type fields struct {
		trie  *Trie
		noise *regexp.Regexp
	}
	type args struct {
		rd io.Reader
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := &Filter{
				trie:  tt.fields.trie,
				noise: tt.fields.noise,
			}
			if err := filter.Load(tt.args.rd); (err != nil) != tt.wantErr {
				t.Errorf("Filter.Load() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
