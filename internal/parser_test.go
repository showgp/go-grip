package internal

import (
	"strings"
	"testing"
)

func TestRenderCollectsTOCEntries(t *testing.T) {
	t.Parallel()

	rendered, err := NewParser().Render([]byte("# Intro\n\n## Setup\n\n### Details\n"))
	if err != nil {
		t.Fatalf("render markdown: %v", err)
	}

	want := []TOCEntry{
		{Level: 1, Text: "Intro", ID: "intro"},
		{Level: 2, Text: "Setup", ID: "setup"},
		{Level: 3, Text: "Details", ID: "details"},
	}
	if len(rendered.TOC) != len(want) {
		t.Fatalf("expected %d TOC entries, got %d: %#v", len(want), len(rendered.TOC), rendered.TOC)
	}
	for i := range want {
		if rendered.TOC[i] != want[i] {
			t.Fatalf("entry %d: expected %#v, got %#v", i, want[i], rendered.TOC[i])
		}
	}
}

func TestRenderTOCUsesGeneratedDuplicateHeadingIDs(t *testing.T) {
	t.Parallel()

	rendered, err := NewParser().Render([]byte("# Intro\n\n# Intro\n"))
	if err != nil {
		t.Fatalf("render markdown: %v", err)
	}

	wantIDs := []string{"intro", "intro-1"}
	if len(rendered.TOC) != len(wantIDs) {
		t.Fatalf("expected %d TOC entries, got %d: %#v", len(wantIDs), len(rendered.TOC), rendered.TOC)
	}
	for i, want := range wantIDs {
		if rendered.TOC[i].ID != want {
			t.Fatalf("entry %d: expected ID %q, got %q", i, want, rendered.TOC[i].ID)
		}
	}
}

func TestRenderStrongWithCJKTextAndParentheses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		input        string
		want         string
		literalStars bool
	}{
		{
			name:  "standalone",
			input: "**大卫·S·邓肯（David S. Duncan）**\n",
			want:  "<strong>大卫·S·邓肯（David S. Duncan）</strong>",
		},
		{
			name:  "followed by cjk text",
			input: "**大卫·S·邓肯（David S. Duncan）**是Innosight的高级合伙人。\n",
			want:  "<strong>大卫·S·邓肯（David S. Duncan）</strong>是Innosight的高级合伙人。",
		},
		{
			name:  "surrounded by cjk text",
			input: "我的合著者**泰迪·霍尔（Taddy Hall）**是我在哈佛商学院的第一堂课上的学生。\n",
			want:  "我的合著者<strong>泰迪·霍尔（Taddy Hall）</strong>是我在哈佛商学院的第一堂课上的学生。",
		},
		{
			name:  "content wrapped in east asian punctuation",
			input: "太郎は**「こんにちわ」**と言った\n",
			want:  "太郎は<strong>「こんにちわ」</strong>と言った",
		},
		{
			name:  "full user sample",
			input: "我的合著者**泰迪·霍尔（Taddy Hall）**是我在哈佛商学院的第一堂课上的学生，多年来我们一直在项目上合作，包括与Intuit创始人斯科特·库克合著了《哈佛商业评论》（*HBR*）文章\"营销失当\"（Marketing Malpractice），该文章首次在《哈佛商业评论》上提出了待办任务理论。他目前是Cambridge Group（Nielsen公司的一部分）的负责人，也是Nielsen突破性创新项目的领导者。因此，他与一些世界领先的公司密切合作，包括本书中提到的许多公司。更重要的是，他多年来一直在他的创新咨询工作中运用任务理论。\n\n**凯伦·狄龙（Karen Dillon）**是《哈佛商业评论》的前任编辑，也是《你将如何衡量你的人生？》（*How Will You Measure Your Life?*）的合著者。你会在本书中看到她在媒体组织担任高级管理人员的长期视角，以及她如何努力把创新做对。在我们的合作过程中，她将自己的角色视为你——读者——的代理。她也是我在连接学术界和实践者世界方面最值得信赖的盟友之一。\n\n**大卫·S·邓肯（David S. Duncan）**是Innosight的高级合伙人，这是一家我在2000年共同创立的咨询公司。他是创新战略和增长领域领先的思想家和高级管理人员顾问，帮助他们驾驭颠覆性变革、创造可持续增长并改造组织以实现长期繁荣。他合作过的客户告诉我，他们完全改变了对业务的思考方式，并将企业文化转变为了真正关注顾客任务的。（有一位客户甚至以他的名字命名了一间会议室。）在过去十年中，他在帮助开发和实施任务理论方面的工作使他成为该领域最富知识和创新精神的实践者之一。\n",
			want:  "<strong>大卫·S·邓肯（David S. Duncan）</strong>是Innosight的高级合伙人",
		},
		{
			name:         "code span remains literal",
			input:        "`**大卫·S·邓肯（David S. Duncan）**是Innosight的高级合伙人。`\n",
			want:         "<code>**大卫·S·邓肯（David S. Duncan）**是Innosight的高级合伙人。</code>",
			literalStars: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rendered, err := NewParser().Render([]byte(tt.input))
			if err != nil {
				t.Fatalf("render markdown: %v", err)
			}

			if !strings.Contains(rendered.Content, tt.want) {
				t.Fatalf("expected strong emphasis %q, got %q", tt.want, rendered.Content)
			}
			if !tt.literalStars && (strings.Contains(rendered.Content, "**大卫") || strings.Contains(rendered.Content, "**泰迪") || strings.Contains(rendered.Content, "**凯伦")) {
				t.Fatalf("expected emphasis delimiters to be consumed, got %q", rendered.Content)
			}
		})
	}
}
