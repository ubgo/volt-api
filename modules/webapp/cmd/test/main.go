package main

import (
	"core/pkg/applog"
	"webapp"

	"github.com/ubgo/errx"
)

func test() *errx.Error {
	return errx.New("test").WithPublic("test2").WithCode("123").WithSafe("data", "test3")
}

func main() {
	_ = webapp.New()

	applog.GetLogger().Info("test")

	// client := circuit.EntDB.Client()
	// ctx := context.Background()
	// goutil.Use(client, ctx)

	// // goerr.AddCallback(func(resp *goerr.ErrorResponse) {
	// // 	go func() {
	// // 		time.Sleep(10 * time.Second)
	// // 		fmt.Println("callback", resp)
	// // 	}()
	// // })

	// err := test()
	// if err != nil {
	// 	goutil.PrintToJSON(err.Response())
	// 	// fmt.Println(err.Response().Code)
	// 	fmt.Println(err.ErrorUser())
	// }

	// // gozap.GetLogger().Info("test")

	// time.Sleep(30 * time.Second)

	// limitfn.CalculateAndSaveWorkspaceFeatureLimitsFinal(ctx, client, "um1yaz1ig0b5504r3w")

	// limitfn.UpdateTotalStorageUsage(ctx, "um1yaz1ig0b5504r3w", client)

	// var orderAggregates struct {
	// 	TotalVecDims      int `json:"total_vec_dims"`
	// 	TotalContentChars int `json:"total_content_chars"`
	// }
	// err := client.Item.Query().
	// 	// Where(item.WorkspaceID("ueqempirgtj02vz4pg1")).
	// 	Aggregate(ent.As(entdb.AggOrZero(ent.Sum(item.FieldContentChars)), "total_content_chars"), ent.As(entdb.AggOrZero(ent.Sum(item.FieldVecDims)), "total_vec_dims")).
	// 	Scan(ctx, &orderAggregates)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// goutil.PrintToJSON(orderAggregates)

	// aicache.GetAiCache().ClearAllCache()

	// billingfn.GetCustomerInWorkspace("ueqempirgtj02vz4pg11", client)
	// _, err := billingfn.VerifyCheckoutSession("cs_test_b1Qg2erznV3yjgIDPCANXleOGLaGoIJS1E7c5MjtubPzJ5NWXzHULxPA6U", client)
	// _, err = billingfn.VerifyCheckoutSession("cs_test_b1X1qGb0n69FbXj8E01GBjg7JP9dm191AaAeNKkpkn9Ctt1h7BlMLWjQsr", client)

	// err := billingfn.FetchAndSyncLast100Invoices(client)

	// err := billingfn.FetchAndSyncInvoice("in_1Rq7IdAq0sBzenJUWJ14tlYL", client)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// charge, err := billingfn.FetchInvoiceByChargeID("py_3Rnw8qAq0sBzenJU1Y2F7eaM", client)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// goutil.PrintToJSON(charge)

	// sc := billingfn.GetClient()

	// invoice, err := sc.V1Invoices.Retrieve(context.Background(), "in_1Rnw8tAq0sBzenJUUyBGdNQk", &stripe.InvoiceRetrieveParams{
	// 	Expand: []*string{
	// 		stripe.String("payments"),
	// 	},
	// })

	// fmt.Println(err)

	// goutil.PrintToJSON(invoice)

	// err := limitfn.UpdateLimitByID(limitfn.FeatureLimitIDChatbot, "ueqempirgtj02vz4pg", 1, limitfn.LimitUpdateTypeAdd, client)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// err := limitfn.CalculateAndSaveWorkspaceFeatureLimitsFinal(ctx, client, "ueqempirgtj02vz4pg1")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// err := limitfn.UpdateWorkspaceUsage(ctx, "ueqempirgtj02vz4pg1", client)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// embeddingProvider, err := appfn.GetEmbeddingProviderFromAIModelID("em2", client)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// embeddings, err := embeddingProvider.GetEmbeddingClient().GetEmbeddings([]string{"hi", "name"}, 10)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// goutil.PrintToJSON(embeddings)

	// fmt.Println("done")

	// _, err := client.Item.Update().Modify(func(u *sql.UpdateBuilder) {
	// 	u.Set(item.FieldName, "t.embedding_msg")
	// 	u.FromSelect(sql.Select().From(sql.Table("tmp_item").As("t")).
	// 		Where(sql.EQ("items.id", "t.item_id"))).
	// 		Where(sql.EQ("t.batch_id", "atmix4ciqbr1eel"))
	// }).Save(ctx)
	// fmt.Println(err)

	// cfg, err := config.Get()  // configuration comes from Go defaults + env
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println(cfg)
	// fmt.Printf("I'm running on host %s\n", cfg.Hostname)

	// // circuit.RedisClient.JSONDel(ctx, "chatbot-ufpdv3nn36ue5jfv9m", "$")
	// // _, err := aifn.GetEmbeddings([]string{"Hello! How can I assist you today", "name"})
	// // if err != nil {
	// // 	fmt.Println(err)
	// // }
	// t := tokenizer.MustCalToken(`

	// Hello! How can I a

	// ssist you today name`)
	// fmt.Println(t) // Output: 64

	// goutil.PrintToJSON(embeddings)

	// cred, err := client.Credential.Create().SetName("test").SetPassword("admin").Save(ctx)
	// fmt.Println(err)
	// goutil.PrintToJSON(cred)

	// fmt.Println(crypt.Decrypt("eaea04d3a7f8f2141d01112021f86b1f86b0d19455e8b798e2c19722a6931f3c"))

	// creds, _ := client.Credential.Query().Where(credential.ID("u5epthhtw6rzixg54a")).First(ctx)
	// goutil.PrintToJSON(creds)
	// var aa string
	// err := creds.Password.Scan(&aa)
	// fmt.Println(err)
	// fmt.Println("sdfds", aa)
	// goutil.PrintToJSON(creds)

	// featurefn.LoadFeatures("ueqempirgtj02vz4pg", client)

	// featureChecker, err := featurefn.NewFeatureChecker(client, "ueqempirgtj02vz4pg")
	// if err != nil {
	// 	log.Fatalf("Error creating feature checker: %v", err)
	// 	return
	// }
	// fmt.Println(featureChecker.Available(featurefn.FeatureNameChatbots))

	// aa := openai.ChatCompletionMessage{
	// 	Role:    openai.ChatCompletionMessageRoleAssistant,
	// 	Content: "Hello!",
	// }
	// goutil.PrintToJSON(aa)

	// err := featurefn.FeaturesSummaryRefresh("ueqempirgtj02vz4pg")
	// fmt.Println(err)
	// textToDcuments()
	// channel, err := client.Channel.Query().WithChannelUsers(func(cuq *ent.ChannelUserQuery) {
	// 	cuq.Where(channeluser.UserIDIn("uhxz8mc4qhmjylgbeb"))
	// 	cuq.WithUser()
	// }).
	// 	Where(
	// 		channel.ID("u565hq36tuc4e113j6"),
	// 		channel.HasChannelUsersWith(channeluser.UserID("uhxz8mc4qhmjylgbeb")),
	// 		channel.HasChannelUsersWith(channeluser.UserID("ussofyhvbggg0xmidv")),
	// 	).First(ctx)
	// fmt.Println(err)
	// goutil.PrintToJSON(channel)

	// featuers1, _ := client.Feature.Query().WithFeatureSummaries().All(ctx)
	// goutil.PrintToJSON(featuers1)

	// textToSplit()
	// insertSampleEmbedding(client)
	// testQueryEmbedding(client)

	// generateEmbeddingFromDir(client)
	// d1 := []byte("hello\ngo\n")
	// err := os.WriteFile("/Volumes/D/www/projects/thewildai/aitools_go/embedding.txt", d1, 0644)
	// fmt.Println(err)

	// return

	// var m ent.ChatbotUserMutation
	// m.OldChatbotID()
	// client.Chatbot.UpdateOneID("uxgcrzd7u8fy8a6ema").SetStatus(true).Save(ctx)
	// client.Chatbot.Create().SetName("test").Save(ctx)

	// clientai := appfn.GetOpenAIClient()

	// result, err := clientai.Embeddings.New(context.Background(), openai.EmbeddingNewParams{
	// 	Input: openai.F[openai.EmbeddingNewParamsInputUnion](shared.UnionString("The quick")),
	// 	Model: openai.F(openai.EmbeddingModelTextEmbedding3Small),
	// 	// Dimensions:     openai.F(int64(1)),
	// 	EncodingFormat: openai.F(openai.EmbeddingNewParamsEncodingFormatFloat),
	// })

	// fmt.Println(err)
	// fmt.Println(result.Data[0].Embedding)

	// goutil.PrintToJSON(result)

	// var embedding []float32 = []float32{0, 1, 2, 3, 4, 5}

	// for _, e := range result.Data[0].Embedding {
	// 	embedding = append(embedding, float32(e))
	// }

	// _, err = client.Item.
	// 	Create().
	// 	// SetEmbedding(pgvector.NewVector([]float32{1, 2, 3})).
	// 	SetEmbedding(pgvector.NewVector(embedding)).
	// 	Save(ctx)

	// fmt.Println(err)

	// Get embeddings for the search string
	// embeddings, err := aifn.GetEmbeddings([]string{"ga4"})
	// if err != nil {
	// 	fmt.Errorf("failed to generate embeddings: %v", err)
	// 	return
	// }

	// items, err := client.Item.
	// 	Query().
	// 	Where(item.ItemGroupID("sdfds")).
	// 	Order(func(s *sql.Selector) {

	// 		s.OrderExpr(sql.ExprP("embedding <-> $2", "[-0.009877119,0.0015331338]"))

	// 		// s.OrderExpr(sql.Expr(fmt.Sprintf("embedding <=> '%s'::vector", "[-0.009877119,0.0015331338]")))
	// 	}).
	// 	Limit(5).
	// 	All(ctx)

	// for _, item := range items {
	// 	fmt.Println(item.ID)
	// }

	// fmt.Println(err)

	// channelRec, _ := client.Channel.Query().Where(channel.ID("ufwbid176cqdtst9cy")).First(ctx)
	// _, err := channelRec.Update().AddUserIDs("u5zlb24k0mkl7o5hln", "uxgcrzd7u8fy8a6ema").Save(ctx)
	// fmt.Println(err)

	// res, err := saasbyteapigql.QueuePullJobs(context.Background(), saasbyteeconf.GetClient(), nil, &saasbyteapigql.QueuePullJobsInput{
	// 	Limit: utils.Ptr(100),
	// })

	// goutil.PrintToJSON(res)
	// fmt.Println(err)

	// res, err := saasbyteapigql.AuthJwtVerify(context.Background(), saasbyteeconf.GetClient(), "")

	// astradb.Test()
	// pineconedb.Run()

	// tools.FetchHtml(toolstypes.HttpRequestOptions{
	// 	URL: "https://www.funfun.tools",
	// })

	// chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
	// 	Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
	// 		openai.UserMessage("Who are you?"),
	// 	}),
	// 	Model: openai.F(openai.ChatModelGPT4o),
	// })
	// if err != nil {
	// 	panic(err.Error())
	// }
	// println(chatCompletion.Choices[0].Message.Content)

	// param := openai.ChatCompletionNewParams{
	// 	Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
	// 		openai.UserMessage("What kind of houseplant is easy to take care of?"),
	// 	}),
	// 	Seed:  openai.Int(1),
	// 	Model: openai.F(openai.ChatModelGPT4o),
	// }

	// completion, err := client.Chat.Completions.New(ctx, param)

	// goutil.PrintToJSON(completion.Choices)

	// // param.Messages.Value = append(param.Messages.Value, completion.Choices[0].Message)
	// // param.Messages.Value = append(param.Messages.Value, openai.UserMessage("How big are those?"))

	// // println(completion.Choices[0].Message.Content)

	// // goutil.PrintToJSON(param.Messages)
	// // // continue the conversation
	// // // completion, err = client.Chat.Completions.New(ctx, param)

	// // // println(completion.Choices[0].Message.Content)
	// fmt.Println(err)

	// body, err := html2json.UrlToJson(html2json.UrlToJsonParams{
	// 	Url: "http://localhost/scrap1.html",
	// })

	// // fmt.Println(body)
	// goutil.PrintToJSON(body)
	// fmt.Println(err)

	// 	input := `
	// 	<!DOCTYPE html>
	// <html>
	// <body>

	// <h1>The script element</h1>

	// <p id="demo">dd</p>

	// <script>document.getElementById("demo").innerHTML = "Hello JavaScript!";</script>

	// </body>
	// </html>
	// `

	// 	buf := new(bytes.Buffer)
	// 	weirdWriter := new(bytes.Buffer)

	// 	debug := io.MultiWriter(buf, weirdWriter) // attach buf to weirdWriter

	// 	aa := graphql.MarshalAny(input)
	// 	aa.MarshalGQL(debug)

	// 	fmt.Println(buf.String())
	// 	fmt.Println(buf.Bytes())

	// markdown, err := htmltomarkdown.ConvertString(input)

	// fmt.Println(err)

	// fmt.Println(markdown)
}
