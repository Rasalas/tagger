package commit

import (
	"testing"

	"github.com/rasalas/tagger/internal/semver"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    Commit
	}{
		{
			name:    "simple feat",
			message: "feat: add login",
			want:    Commit{Type: "feat", Summary: "add login"},
		},
		{
			name:    "feat with scope",
			message: "feat(auth): add login",
			want:    Commit{Type: "feat", Scope: "auth", Summary: "add login"},
		},
		{
			name:    "breaking with bang",
			message: "feat!: remove API",
			want:    Commit{Type: "feat", Breaking: true, Summary: "remove API"},
		},
		{
			name:    "breaking with scope and bang",
			message: "refactor(core)!: rewrite engine",
			want:    Commit{Type: "refactor", Scope: "core", Breaking: true, Summary: "rewrite engine"},
		},
		{
			name:    "breaking in body",
			message: "feat: change API\n\nBREAKING CHANGE: old endpoint removed",
			want:    Commit{Type: "feat", Breaking: true, Summary: "change API", Body: "BREAKING CHANGE: old endpoint removed"},
		},
		{
			name:    "breaking-change in body",
			message: "feat: change API\n\nBREAKING-CHANGE: old endpoint removed",
			want:    Commit{Type: "feat", Breaking: true, Summary: "change API", Body: "BREAKING-CHANGE: old endpoint removed"},
		},
		{
			name:    "fix",
			message: "fix: resolve crash",
			want:    Commit{Type: "fix", Summary: "resolve crash"},
		},
		{
			name:    "non-conventional",
			message: "update readme",
			want:    Commit{Summary: "update readme"},
		},
		{
			name:    "uppercase type is case-insensitive",
			message: "Feat: add login",
			want:    Commit{Type: "feat", Summary: "add login"},
		},
		{
			name:    "mixed-case type is lowercased",
			message: "FIX(ui): resolve crash",
			want:    Commit{Type: "fix", Scope: "ui", Summary: "resolve crash"},
		},
		{
			name:    "unclosed scope is not conventional",
			message: "feat(auth: add login",
			want:    Commit{Summary: "feat(auth: add login"},
		},
		{
			name:    "unclosed scope with bang does not stay breaking",
			message: "feat(auth!: add login",
			want:    Commit{Summary: "feat(auth!: add login"},
		},
		{
			name:    "chore",
			message: "chore: update deps",
			want:    Commit{Type: "chore", Summary: "update deps"},
		},
		{
			name:    "with body",
			message: "fix(ui): button color\n\nThe button was blue instead of green.",
			want:    Commit{Type: "fix", Scope: "ui", Summary: "button color", Body: "The button was blue instead of green."},
		},
		{
			name:    "gitlab merge branch feat",
			message: "Merge branch 'feat/split-gitlab-ci' into 'main'",
			want:    Commit{Type: "feat", Summary: "split-gitlab-ci"},
		},
		{
			name:    "gitlab merge branch fix with nested path",
			message: "Merge branch 'fix/gitlab/pipeline-status' into 'main'",
			want:    Commit{Type: "fix", Summary: "gitlab/pipeline-status"},
		},
		{
			name:    "gitlab merge branch scoped breaking",
			message: "Merge branch 'refactor(core)!/rewrite-engine' into 'main'",
			want:    Commit{Type: "refactor", Scope: "core", Breaking: true, Summary: "rewrite-engine"},
		},
		{
			name:    "gitlab merge branch keeps body",
			message: "Merge branch 'feat/api-v2' into 'main'\n\nSee merge request project/repo!7",
			want:    Commit{Type: "feat", Summary: "api-v2", Body: "See merge request project/repo!7"},
		},
		{
			name:    "gitlab merge branch without conventional branch",
			message: "Merge branch 'release' into 'main'",
			want:    Commit{Summary: "Merge branch 'release' into 'main'"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Parse(tt.message)
			if got.Type != tt.want.Type {
				t.Errorf("Type = %q, want %q", got.Type, tt.want.Type)
			}
			if got.Scope != tt.want.Scope {
				t.Errorf("Scope = %q, want %q", got.Scope, tt.want.Scope)
			}
			if got.Breaking != tt.want.Breaking {
				t.Errorf("Breaking = %v, want %v", got.Breaking, tt.want.Breaking)
			}
			if got.Summary != tt.want.Summary {
				t.Errorf("Summary = %q, want %q", got.Summary, tt.want.Summary)
			}
			if got.Body != tt.want.Body {
				t.Errorf("Body = %q, want %q", got.Body, tt.want.Body)
			}
		})
	}
}

func TestClassify(t *testing.T) {
	tests := []struct {
		name string
		c    Commit
		want semver.BumpLevel
	}{
		{"breaking", Commit{Type: "feat", Breaking: true}, semver.Major},
		{"feat", Commit{Type: "feat"}, semver.Minor},
		{"fix", Commit{Type: "fix"}, semver.Patch},
		{"refactor", Commit{Type: "refactor"}, semver.Patch},
		{"perf", Commit{Type: "perf"}, semver.Patch},
		{"revert", Commit{Type: "revert"}, semver.Patch},
		{"chore", Commit{Type: "chore"}, semver.None},
		{"docs", Commit{Type: "docs"}, semver.None},
		{"empty type", Commit{}, semver.None},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Classify(tt.c); got != tt.want {
				t.Errorf("Classify() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnalyze(t *testing.T) {
	t.Run("mixed commits", func(t *testing.T) {
		messages := []string{
			"fix: patch bug",
			"feat: new feature",
			"chore: update deps",
		}
		level, commits := Analyze(messages)
		if level != semver.Minor {
			t.Errorf("level = %v, want Minor", level)
		}
		if len(commits) != 3 {
			t.Errorf("len(commits) = %d, want 3", len(commits))
		}
	})

	t.Run("breaking wins", func(t *testing.T) {
		messages := []string{
			"fix: patch bug",
			"feat!: breaking change",
		}
		level, _ := Analyze(messages)
		if level != semver.Major {
			t.Errorf("level = %v, want Major", level)
		}
	})

	t.Run("gitlab merge branch contributes bump", func(t *testing.T) {
		messages := []string{
			"Merge branch 'fix/pipeline-status' into 'main'",
			"Merge branch 'feat/split-gitlab-ci' into 'main'",
			"chore: update deps",
		}
		level, commits := Analyze(messages)
		if level != semver.Minor {
			t.Errorf("level = %v, want Minor", level)
		}
		if len(commits) != 3 {
			t.Errorf("len(commits) = %d, want 3", len(commits))
		}
	})

	t.Run("no conventional commits", func(t *testing.T) {
		messages := []string{
			"update readme",
			"misc changes",
		}
		level, _ := Analyze(messages)
		if level != semver.None {
			t.Errorf("level = %v, want None", level)
		}
	})

	t.Run("empty list", func(t *testing.T) {
		level, commits := Analyze(nil)
		if level != semver.None {
			t.Errorf("level = %v, want None", level)
		}
		if len(commits) != 0 {
			t.Errorf("len(commits) = %d, want 0", len(commits))
		}
	})
}
