package main

import (
	"context"
	"fmt"
	"os"

	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/ubgo/goutil"
)

func textToDcuments() {
	f, err := os.Open("/Volumes/D/www/projects/thewildai/aitools_go/uploads/chatbots/uxgcrzd7u8fy8a6ema/1741201967410452000_shopify schema 2024-10.txt")
	if err != nil {
		fmt.Println("Error opening file: ", err)
		return
	}

	p := documentloaders.NewText(f)

	split := textsplitter.NewRecursiveCharacter()
	split.ChunkSize = 1000  // size of the chunk is number of characters
	split.ChunkOverlap = 30 // overlap is the number of characters that the chunks overlap
	docs, err := p.LoadAndSplit(context.Background(), split)

	if err != nil {
		fmt.Println("Error loading document: ", err)
		return
	}

	goutil.PrintToJSON(docs)
}
