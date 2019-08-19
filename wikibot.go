package main

import (
	"encoding/csv"
	"log"
	"net"
	"os"

	"./client"
)

const host = "localhost:3000"

func main() {
	cfg, err := os.Open("config.csv")
	if err != nil {
		log.Panicln(err)
	}
	csvreader := csv.NewReader(cfg)

	values, err := csvreader.Read()
	if err != nil {
		log.Panicln(err)
	}

	botUsername := values[0]
	botPassword := values[1]

	bot := client.Wiki(botUsername, botPassword)

	bot.CompareTranslations("Deadbeats/pt-br")

	// article := []byte(`{{Item infobox
	// 	| type           = cosmetic
	// 	| image          = Deadbeats.png
	// 	| used-by        = [[Classes/ja|全てのクラス]]
	// 	| equip-region   = ears
	// 	| contributed-by = {{Backpack Item Link|76561197972481083|7259562517}}<br/>{{Backpack Item Link|76561197976113694|7247343703}}<br/>{{Backpack Item Link|76561198025795825|7259562125}}
	// 	| released       = {{Patch name|10|19|2018}}<br>([[Scream Fortress 2018/ja|Scream Fortress 2018]])
	// 	| availability   = {{avail|crate120}}
	// 	| trade          = yes
	// 	| gift           = yes
	// 	| paint          = yes
	// 	| rename         = yes
	// 	| numbered       = no
	// 	| loadout        = yes
	// 	  | item-kind    = Headphones
	// 	  | grade        = Mercenary
	// 	}}

	// 	{{Quotation|'''{{item name|Deadbeats}}'''の宣伝文句|このゾッとするけどスタイリッシュなヘッドホンを持ち込んで、墓地でレイヴを開いちゃおう。往年の名曲で、古いガイコツたちもガタガタビートにノっちゃうこと100%間違いナシ。}}

	// 	'''{{item name|Deadbeats}}'''は[[Steam Workshop/ja|コミュニティ製]]の[[Classes/ja|全クラス]]用[[Cosmetic items/ja|装飾アイテム]]です。これはペイント可能なドクロがついた黒いヘッドホンです。

	// 	このアイテムには2つの[[styles/ja|スタイル]]があり、"Hat" はデフォルトの帽子の上から、"No Hat" ではデフォルトのヘッドギアを[[Bodygroup/ja|外して]]ヘッドホンを装着します。デフォルトのヘッドギアを持たないクラスでは、スタイルを変えても外見は変わりません。

	// 	{{item name|Deadbeats}}はSteamワークショップに[https://steamcommunity.com/sharedfiles/filedetails/?id=1179501744 投稿された作品]です。

	// 	== {{common string|Painted variants}} ==
	// 	{{Painted variants}}

	// 	== {{common string|Styles}} ==
	// 	{{Styles table
	// 	| image1 = Deadbeats Hat.png
	// 	| image2 = Deadbeats No Hat.png
	// 	| style1 = Hat
	// 	| style2 = No Hat
	// 	}}

	// 	== {{common string|Update history}} ==
	// 	'''{{Patch name|10|19|2018}}''' ([[Scream Fortress 2018/ja|Scream Fortress 2018]])
	// 	* {{item name|DeadBeats}}がゲームに追加された。

	// 	'''{{Patch name|12|19|2018}}''' ([[Smissmas 2018/ja|Smissmas 2018]])
	// 	* {{Undocumented}}{{item name|DeadBeats}}に[[Styles/ja|スタイル]]が追加された。

	// 	== {{common string|Trivia}} ==
	// 	* このアイテム名の元ネタは、ヘッドホンブランドの[[w:ja:ビーツ・エレクトロニクス|''Beats'']]です。
	// 	** ''Deadbeat''とは怠惰な人、無責任な人、評判の悪い人を指す言葉でもあります。

	// 	== {{common string|Gallery}} ==
	// 	<gallery perrow=3>
	// 	File:Scout Deadbeats Hat.png|[[スカウト]]
	// 	File:Soldier Deadbeats Hat.png|[[ソルジャー]]
	// 	File:Pyro Deadbeats Hat.png|[[パイロ]]
	// 	File:Demoman Deadbeats Hat.png|[[デモマン]]
	// 	File:Heavy Deadbeats Hat.png|[[ヘビー]]
	// 	File:Engineer Deadbeats Hat.png|[[エンジニア]]
	// 	File:Medic Deadbeats Hat.png|[[メディック]]
	// 	File:Sniper Deadbeats Hat.png|[[スナイパー]]
	// 	File:Spy Deadbeats Hat.png|[[スパイ]]
	// 	File:Steamworkshop deadbeats thumb.jpg|Steamワークショップでの{{item name|DeadBeats}}のサムネイル画像
	// 	</gallery>

	// 	{{Scream Fortress 2018 Nav}}
	// 	{{Hat Nav}}

	// 	[[Category:Items with styles/ja]]
	// 	`)

	// log.Println(bot.GetLinks(article, "ja"))

	listen, err := net.Listen("tcp", host)

	if err != nil {
		log.Panicf("Error starting up TCP server: %s", err)
	}
	defer listen.Close()
	log.Printf("Server listening at %s", host)

	for {
		connection, err := listen.Accept()
		if err != nil {
			log.Printf("Server error right here blud: %s", err)
		}

		go serverHandle(connection)
	}
}

func serverHandle(connection net.Conn) {
	buffer := make([]byte, 2048)
	requestLength, err := connection.Read(buffer)
	if err != nil {
		log.Panicf("Error handling request: %s", requestLength)
	}
	connection.Write([]byte("i gotchu fam thx bye"))
	connection.Close()
}
