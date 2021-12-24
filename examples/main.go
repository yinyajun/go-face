package main

import (
	"flag"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	//face "github.com/Kagami/go-face"
	face "github.com/yinyajun/go-face"
)

const dataDir = "/data/asset"

var RecognizerCount = 5
var GlobalCount = uint64(0)
var RecognizerList []*face.Recognizer
var images []string
var TestNum int

func init() {
	for i := 0; i < RecognizerCount; i += 1 {
		rec, err := face.NewRecognizer(filepath.Join(dataDir, "models"))
		if err != nil {
			log.Panic("Can't inits face recognizer", err)
		}
		RecognizerList = append(RecognizerList, rec)
	}
	log.Println("Init ok", len(RecognizerList))
	images, _ = ListDir(filepath.Join(dataDir, "images"), "jpg")
	log.Println("Find", len(images), "test images")
	rand.Seed(time.Now().Unix())

	flag.IntVar(&TestNum, "n", 10, "test num")
	flag.Parse()
}

func TestFace(num int) {
	var wg sync.WaitGroup
	for i := 0; i < num; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			index := atomic.AddUint64(&GlobalCount, 1)
			index = index % uint64(len(RecognizerList))
			Recognizer := RecognizerList[index]
			img := images[rand.Intn(len(images))]
			faces, err := Recognizer.RecognizeFile(img)
			if err != nil {
				log.Println(index, img, err)
			} else {
				if len(faces) == 0 {
					log.Println(index, img, len(faces))
					return
				}
				log.Println(index, img, len(faces), faces[0].Rectangle, faces[0].Shapes[32])
			}
		}(i)
	}
	wg.Wait()
}

func main() {
	TestFace(TestNum)
}

func ListDir(dirPth string, suffix string) (files []string, err error) {
	files = make([]string, 0, 10)

	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}

	PthSep := string(os.PathSeparator)
	suffix = strings.ToUpper(suffix)

	for _, fi := range dir {
		if fi.IsDir() {
			continue
		}
		if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) {
			files = append(files, dirPth+PthSep+fi.Name())
		}
	}
	return files, nil
}
