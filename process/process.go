package process

import (
	"SpeechEnhancement/media_info"
	"SpeechEnhancement/meta"
	"SpeechEnhancement/upload"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

const abortIndex int = 3

type Context struct {
	index int
	Keys  map[string]any
	mu    sync.RWMutex
	meta  meta.Meta
}

func (c *Context) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Keys == nil {
		c.Keys = make(map[string]any)
	}

	c.Keys[key] = value
}
func (c *Context) Get(key string) (value any, exists bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, exists = c.Keys[key]
	return
}
func (c *Context) Abort() {
	c.index = abortIndex
	puuid, _ := c.Get("uuid")
	err := meta.Set(c.meta, puuid.(string), meta.KeyStatus, meta.StatusError)
	if err != nil {
		log.Printf("Error:Transcode Meta Set %s\n", err)
		return
	}
}

type Process struct {
	UUID      string
	Source    string
	MediaInfo media_info.AudioInfoRes
	Context   Context
	TargetDir string
	Chains    []func(c *Context)
}

func New(uuid string, source string) (*Process, error) {
	dir := filepath.Join(upload.DefaultUploadPath, uuid)
	sourcePath := filepath.Join(dir, source)
	mediaInfo, err := media_info.AudioInfo(sourcePath)
	if err != nil {
		log.Printf("Error:NewProcess %s\n", err)
		return &Process{}, err
	}
	var m meta.M

	return &Process{
		UUID:      uuid,
		Source:    sourcePath,
		MediaInfo: mediaInfo,
		TargetDir: dir,
		Context: Context{
			index: 0,
			Keys:  map[string]any{},
			meta:  m,
		},
	}, nil
}

func (p *Process) SetChain() *Process {
	if p.MediaInfo.Format.FormatName != "wav" || p.MediaInfo.AudioStream[0].SampleRate != "44100" || p.MediaInfo.AudioStream[0].BitsPerSample != 16 {
		p.Chains = append(p.Chains, transcode)
	}
	p.Chains = append(p.Chains, split)
	p.Chains = append(p.Chains, batchEnhance)
	p.Chains = append(p.Chains, batchNormalize)
	p.Chains = append(p.Chains, merge)
	p.Chains = append(p.Chains, clear)
	return p
}

func (p *Process) Handle() {
	p.Context.Set("source", p.Source)
	p.Context.Set("targetDir", p.TargetDir)
	p.Context.Set("uuid", p.UUID)
	for p.Context.index < len(p.Chains) {
		p.Chains[p.Context.index](&p.Context)
		p.Context.index++
	}
	status, err := meta.Get(p.Context.meta, p.UUID, meta.KeyStatus)
	if err != nil {
		log.Printf("Error:Handle Meta Get %s\n", err)
		return
	}
	if status != meta.StatusError {
		err = meta.Set(p.Context.meta, p.UUID, meta.KeyStatus, meta.StatusComplete)
		if err != nil {
			log.Printf("Error:Handle Meta Set %s\n", err)
			return
		}
	}
}

func transcode(c *Context) {
	puuid, _ := c.Get("uuid")
	err := meta.Set(c.meta, puuid.(string), meta.KeyStatus, meta.StatusTranscode)
	if err != nil {
		log.Printf("Error:Transcode Meta Set %s\n", err)
		c.Abort()
		return
	}

	targetDir, _ := c.Get("targetDir")
	targetFile := filepath.Join(targetDir.(string), "transcode.wav")
	sourceFile, _ := c.Get("source")
	cmd := exec.Command("ffmpeg", "-y", "-v", "quiet", "-hide_banner", "-i", sourceFile.(string), "-acodec", "pcm_s16le", "-ar", "44100", targetFile)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error:Transcode Command %s -- %s\n", err, out)
		c.Abort()
		return
	}
	c.Set("source", targetFile)

	return
}

func split(c *Context) {
	sourceFile, _ := c.Get("source")
	targetDir, _ := c.Get("targetDir")
	partOutDir := filepath.Join(targetDir.(string), "parts")
	if err := os.MkdirAll(partOutDir, 0755); err != nil {
		log.Printf("Error:Split %s\n", err)
		c.Abort()
		return
	}
	cmd := exec.Command("ffmpeg", "-y", "-v", "quiet", "-hide_banner", "-i", sourceFile.(string), "-vn", "-acodec", "copy", "-segment_time", "300", "-f", "segment", filepath.Join(partOutDir, "%03d.wav"))
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error:Split Command %s -- %s\n", err, out)
		c.Abort()
		return
	}
	c.Set("source", partOutDir)
	return
}

func batchEnhance(c *Context) {
	puuid, _ := c.Get("uuid")
	err := meta.Set(c.meta, puuid.(string), meta.KeyStatus, meta.StatusEnhance)
	if err != nil {
		log.Printf("Error:BatchEnhance Meta Set %s\n", err)
		c.Abort()
		return
	}

	sourceDir, _ := c.Get("source")
	var files []string
	err = filepath.Walk(sourceDir.(string), func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		log.Printf("Error:BatchEnhance %s\n", err)
		c.Abort()
		return
	}
	var wg sync.WaitGroup
	for _, file := range files {
		wg.Add(1)
		go func(file string) {
			defer wg.Done()
			enhance(c, file)
		}(file)
	}
	wg.Wait()
	targetDir, _ := c.Get("targetDir")
	enhanceOutDir := filepath.Join(targetDir.(string), "enhance")
	c.Set("source", enhanceOutDir)

}

func enhance(c *Context, sourceFile string) {
	targetDir, _ := c.Get("targetDir")
	enhanceOutDir := filepath.Join(targetDir.(string), "enhance")
	cmd := exec.Command("deep-filter", "-o", enhanceOutDir, sourceFile)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error:Enhance Command %s -- %s\n", err, out)
		c.Abort()
		return
	}
	return
}

func batchNormalize(c *Context) {
	puuid, _ := c.Get("uuid")
	err := meta.Set(c.meta, puuid.(string), meta.KeyStatus, meta.StatusNormalize)
	if err != nil {
		log.Printf("Error:BatchNormalize Meta Set %s\n", err)
		c.Abort()
		return
	}

	var files []string
	sourceDir, _ := c.Get("source")
	err = filepath.Walk(sourceDir.(string), func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		log.Printf("Error:BatchNormalize Command %s\n", err)
		c.Abort()
		return
	}
	targetDir, _ := c.Get("targetDir")
	normalizeOutDir := filepath.Join(targetDir.(string), "normalize")
	if err = os.MkdirAll(normalizeOutDir, 0755); err != nil {
		log.Printf("Error:BatchNormalize %s\n", err)
		c.Abort()
		return
	}
	var wg sync.WaitGroup
	for _, file := range files {
		wg.Add(1)
		go func(file string) {
			defer wg.Done()
			normalize(c, file)
		}(file)
	}
	wg.Wait()

	c.Set("source", normalizeOutDir)
	return
}

func normalize(c *Context, sourceFile string) {
	targetDir, _ := c.Get("targetDir")
	_, file := filepath.Split(sourceFile)
	fileParts := strings.Split(file, ".")
	file = strings.Join(fileParts[:len(fileParts)-1], ".") + ".wav"
	targetFile := filepath.Join(targetDir.(string), "normalize", file)
	cmd := exec.Command("ffmpeg-normalize", sourceFile, "-ar", "44100", "-t", "-15", "-o", targetFile)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error:Normalize Command %s -- %s\n", err, out)
		c.Abort()
		return
	}
	return
}

func merge(c *Context) {
	sourceDir, _ := c.Get("source")
	targetDir, _ := c.Get("targetDir")
	outFile := filepath.Join(targetDir.(string), "out.wav")

	var files []string
	err := filepath.Walk(sourceDir.(string), func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		log.Printf("Error:Merge %s\n", err)
	}
	sort.Strings(files)
	var args []string
	for _, file := range files {
		args = append(args, "-i", file)
	}
	args = append(args, "-filter_complex")

	st := ""
	for i := 0; i < len(files); i++ {
		st += fmt.Sprintf("[%d:0]", i)
	}
	st = st + "concat=n=" + fmt.Sprintf("%d", len(files)) + ":v=0:a=1[out]"
	args = append(args, st)
	args = append(args, "-map", "[out]", outFile)
	cmd := exec.Command("ffmpeg", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error:Merge Command %s -- %s\n", err, out)
		c.Abort()
		return
	}
	puuid, _ := c.Get("uuid")
	err = meta.Set(c.meta, puuid.(string), meta.KeyTarget, outFile)
	if err != nil {
		log.Printf("Error:Merge Meta Set %s\n", err)
		c.Abort()
		return
	}
}

func clear(c *Context) {
	targetDir, _ := c.Get("targetDir")
	puuid, _ := c.Get("uuid")
	sourceName, err := meta.Get(c.meta, puuid.(string), meta.KeySource)
	if err != nil {
		log.Printf("Error:Merge Meta Set %s\n", err)
		c.Abort()
		return
	}
	sourceFile := filepath.Join(targetDir.(string), sourceName.(string))
	partsOutDir := filepath.Join(targetDir.(string), "parts")
	enhanceOutDir := filepath.Join(targetDir.(string), "enhance")
	normalizeOutDir := filepath.Join(targetDir.(string), "normalize")
	os.Remove(sourceFile)
	os.RemoveAll(partsOutDir)
	os.RemoveAll(enhanceOutDir)
	os.RemoveAll(normalizeOutDir)
	return
}
