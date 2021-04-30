package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"runtime"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

type response struct {
	Found bool        `json:"found"`
	Seed  interface{} `json:"seed"`
	Hash  string      `json:"hash"`
}

func (h *HashArray) getHashes() {
	gets256 := func(s string) string {
		dig := sha256.Sum256([]byte(s))
		return hex.EncodeToString(dig[:])
	}
	ha := *h
	for i := 1; i < len(ha); i++ {
		ha[i] = gets256(ha[i-1])
	}
}

func (h *HashArray) getLastItem() string {
	return (*h)[len(*h)-1]
}

func h2b(s string) string {
	b, _ := hex.DecodeString(s)
	return base64.RawStdEncoding.EncodeToString(b)
}

func b2h(s string) string {
	b, _ := base64.RawStdEncoding.DecodeString(s)
	return hex.EncodeToString(b)
}

func (h *HashArray) hex2b64() {
	for i := 0; i < len(*h); i++ {
		if len((*h)[i]) != 64 {
			continue
		}
		(*h)[i] = h2b((*h)[i])
	}
}

func main() {
	runtime.GOMAXPROCS(128)

	var (
		redis         *Client
		foundFile     *Fs
		newHashesFile *Fs
		err           error
		foundPath     string = "/home/node/found.csv"
		newPath       string = "/home/node/newhashes.csv"
	)

	if redis, err = NewRedisClient(); err != nil {
		log.Fatal("Couldnt connect to redis instance", err)
	}
	defer redis.client.Close()

	if foundFile, err = FileOpen(foundPath); err != nil {
		log.Fatal("Couldnt open file", foundPath, err)
	}
	defer foundFile.CloseFile()

	if newHashesFile, err = FileOpen(newPath); err != nil {
		log.Fatal("Couldnt open file", foundPath, err)
	}
	defer newHashesFile.CloseFile()

	app := fiber.New(fiber.Config{
		Prefork: true,
	})
	app.Use(cors.New())

	app.Get("api/getdbsize", func(c *fiber.Ctx) error {
		dbsize, err := redis.client.DBSize(redis.client.Context()).Result()
		if err != nil {
			return c.Next()
		}
		return c.JSON(dbsize * 1000000)
	})

	app.Get("api/million/:id", func(c *fiber.Ctx) error {
		length := 1000000 + 1
		firstValue := c.Params("id")

		h := make(HashArray, length)
		h[0] = firstValue
		h.getHashes()
		h.hex2b64()

		lastValue := h.getLastItem()
		foundVal, err := redis.GetData(&h)
		if err != nil {
			return c.Next()
		}

		if foundVal != nil && foundVal != h2b(firstValue) {
			go foundFile.Write2File(fmt.Sprintf("seed: %v, hash: %v, lastItem: %v", foundVal, firstValue, lastValue))
			return c.JSON(&response{true, foundVal, firstValue})
		}

		go newHashesFile.Write2File(firstValue + "," + b2h(fmt.Sprintf("%v", lastValue)))
		return c.JSON(&response{false, "", firstValue})
	})
	app.Use(func(c *fiber.Ctx) error {
		return c.SendStatus(404)
	})
	log.Fatal(app.Listen(":3000"))
}
