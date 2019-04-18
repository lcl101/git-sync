package core

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"time"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

var (
	files map[string]bool
)

//Config 配置
type Config struct {
	SrcPath    string   `json:"srcPath"`
	DstPath    string   `json:"dstPath"`
	Type       int8     `json:"type"`
	CommitDate Time     `json:"commitDate"`
	Author     string   `json:"author"`
	Logs       []string `json:"logs"`
	Sync       bool     `json:"sync"`
}

//App app结构
type App struct {
	ConfigPath string
	config     Config
}

func (app *App) String() string {
	return "srcPath: " + app.config.SrcPath + "\ndstPath: " + app.config.DstPath
}

//LoadConfig 加载配置文件
func (app *App) LoadConfig() {
	b, _ := ioutil.ReadFile(app.ConfigPath)
	err := json.Unmarshal(b, &app.config)
	if err != nil {
		Warning("加载配置文件失败", err)
		panic(errors.New("加载配置文件失败：" + err.Error()))
	}
}

//Sync 运行同步
func (app *App) Sync() {
	files = make(map[string]bool)
	if app.config.Type == 1 {
		app.syncByLogs()
	} else if app.config.Type == 2 {
		app.syncByCommitTime()
	}
	app.copy()
}

func (app *App) syncByLogs() {
	r, err := git.PlainOpen(app.config.SrcPath)
	CheckIfError(err)

	for _, s := range app.config.Logs {
		h := plumbing.NewHash(s)
		commit, err := r.CommitObject(h)
		CheckIfError(err)
		app.dealCommit(commit)
	}
}

func (app *App) syncByCommitTime() {
	r, err := git.PlainOpen(app.config.SrcPath)
	CheckIfError(err)
	ref, err := r.Head()
	CheckIfError(err)
	cIter, err := r.Log(&git.LogOptions{From: ref.Hash()})
	CheckIfError(err)
	// ... just iterates over the commits, printing it
	err = cIter.ForEach(func(c *object.Commit) error {
		if c.Committer.When.Unix() >= time.Time(app.config.CommitDate).Unix() {
			app.dealCommit(c)
		}
		return nil
	})
}

func (app *App) dealCommit(commit *object.Commit) {
	if commit.Author.Name != app.config.Author {
		return
	}
	//根据提交消息剔除merge内容
	msg := commit.Message
	if strings.HasPrefix(msg, "Merge branch") {
		Warning("merge commit: %s", msg)
		return
	}

	fs, err := commit.Stats()
	CheckIfError(err)
	// fmt.Println(fs)
	if len(fs) > 0 {
		Info("%s===>%s", commit.Author.Name, commit.Message)
	}
	for _, f := range fs {
		if _, ok := files[f.Name]; !ok {
			files[f.Name] = true
		}
	}

}

func (app *App) copy() {
	if !app.config.Sync {
		return
	}
	for k, v := range files {
		if v {
			src := app.config.SrcPath + "/" + k
			if !CheckFile(src) {
				Error(src)
				continue
			}
			dst := app.config.DstPath + "/" + k
			//删除原文件
			os.Remove(dst)
			size, err := CopyFile(dst, src)
			CheckIfError(err)
			Debug("copy file=%s, size=%d", k, size)
		}
	}
}
