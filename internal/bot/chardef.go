package bot

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v6"
)

func (b *Bot) chardefPullFromRemote(ctx context.Context) error {
	b.logger.Infow("从远程拉取chardefs")

	fp, err := b.pthResolve("chardefs_remote")
	if err != nil {
		return err
	}

	b.logger.Infow("将会拉取到",
		"folder path", fp)

	repo, err := git.PlainOpen(fp)
	if errors.Is(err, git.ErrRepositoryNotExists) {
		// 尝试克隆
		b.logger.Infow("尝试克隆")

		repo, err = git.PlainCloneContext(ctx, fp, &git.CloneOptions{
			URL: "https://github.com/mihari-bot/chardef.git",
		})
		if err != nil {
			return err
		}
	}

	wt, err := repo.Worktree()
	if err != nil {
		return err
	}

	err = wt.PullContext(ctx, &git.PullOptions{Force: true})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return err
	}

	head, err := repo.Head()
	if err != nil {
		return err
	}

	err = wt.Reset(&git.ResetOptions{Mode: git.HardReset, Commit: head.Hash()})
	if err != nil {
		return err
	}

	b.logger.Infow("从远程拉取chardefs完成")

	return nil
}

// ChardefLoadFromFolder 从指定目录加载chardef
func (b *Bot) ChardefLoadFromFolder(pth string) error {
	entries, err := os.ReadDir(pth)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := filepath.Ext(name)
		if ext != ".yaml" {
			continue
		}
		rolename := strings.TrimSuffix(name, ext)

		contentInBytes, err := os.ReadFile(filepath.Join(pth, name))
		if err != nil {
			return err
		}
		content := string(contentInBytes)

		b.chardefMp.Set(rolename, content)
		b.logger.Debugw("chardef loaded",
			"rolename", rolename)
	}

	return nil
}
