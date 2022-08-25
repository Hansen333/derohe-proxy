package proxy

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"time"

	"derohe-proxy/config"

	"github.com/gorilla/websocket"
)

type (
	GetBlockTemplate_Params struct {
		Wallet_Address string `json:"wallet_address"`
		Block          bool   `json:"block"`
		Miner          string `json:"miner"`
	}
	GetBlockTemplate_Result struct {
		JobID              string `json:"jobid"`
		Blocktemplate_blob string `json:"blocktemplate_blob,omitempty"`
		Blockhashing_blob  string `json:"blockhashing_blob,omitempty"`
		Difficulty         string `json:"difficulty"`
		Difficultyuint64   uint64 `json:"difficultyuint64"`
		Height             uint64 `json:"height"`
		Prev_Hash          string `json:"prev_hash"`
		EpochMilli         uint64 `json:"epochmilli"`
		Blocks             uint64 `json:"blocks"`     // number of blocks found
		MiniBlocks         uint64 `json:"miniblocks"` // number of miniblocks found
		Rejected           uint64 `json:"rejected"`   // reject count
		LastError          string `json:"lasterror"`  // last error
		Status             string `json:"status"`
		Orphans            uint64 `json:"orphans"`
		Hansen33Mod        bool   `json:"hansen33_mod"`
	}
)

var connection *websocket.Conn
var Blocks uint64
var Minis uint64
var Rejected uint64
var Orphans uint64
var ModdedNode bool = false
var Hashrate float64

var ReconnectCount = int(0)

var ProxyStart = time.Now()

var old_block_count uint64

func setBlocks(count uint64) {

	if count == 0 {
		return
	}

	if count < old_block_count {
		old_block_count = count
		Blocks += count
		return
	}

	if count > old_block_count {

		Blocks += (count - old_block_count)
		old_block_count = count
	}

}

var old_minis_count uint64

func setMinis(count uint64) {

	if count == 0 {
		return
	}

	if count < old_minis_count {
		old_minis_count = count
		Minis += count
		return
	}

	if count > old_minis_count {

		Minis += (count - old_minis_count)
		old_minis_count = count
	}

}

var old_orphans_count uint64

func setOrphans(count uint64) {

	if count == 0 {
		return
	}

	if count < old_orphans_count {
		old_orphans_count = count
		Orphans += count
		return
	}

	if count > old_orphans_count {

		Orphans += (count - old_orphans_count)
		old_orphans_count = count
	}

}

var old_rejected_count uint64

func setRejected(count uint64) {

	if count == 0 {
		return
	}

	if count < old_rejected_count {
		old_rejected_count = count
		Rejected += count
		return
	}

	if count > old_rejected_count {

		Rejected += (count - old_rejected_count)
		old_rejected_count = count
	}

}

// proxy-client
func Start_client(w string) {
	var err error
	var last_diff uint64
	var last_height uint64

	rand.Seed(time.Now().UnixMilli())

	var error_threshold = int(0)

	for {

		u := url.URL{Scheme: "wss", Host: config.Daemon_address, Path: "/ws/" + w}

		dialer := websocket.DefaultDialer
		dialer.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}

		// if !config.Pool_mode {
		// 	fmt.Printf("%v Connected to node %v\n", time.Now().Format(time.Stamp), config.Daemon_address)
		// } else {
		// 	fmt.Printf("%v Connected to node %v using wallet %v\n", time.Now().Format(time.Stamp), config.Daemon_address, w)
		// }
		connection, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			time.Sleep(100 * time.Microsecond)
			// fmt.Println(err)
			error_threshold++

			if error_threshold > 100 {
				fmt.Print("Too Many Connection Error!\n")
				os.Exit(0)
			}
			continue
		}
		ReconnectCount++

		var params GetBlockTemplate_Result

		for {
			msg_type, recv_data, err := connection.ReadMessage()
			if err != nil {
				// fmt.Printf("\nError: %s\n", err.Error())
				break
			}

			if msg_type != websocket.TextMessage {
				continue
			}

			if err = json.Unmarshal(recv_data, &params); err != nil {
				continue
			}

			error_threshold = 0
			setBlocks(params.Blocks)
			setMinis(params.MiniBlocks)
			setRejected(params.Rejected)
			setOrphans(params.Orphans)

			// if ModdedNode != params.Hansen33Mod {
			// 	if params.Hansen33Mod {
			// 		fmt.Printf("%v Hansen33 Mod Mining Node Detected - Happy Mining\n", time.Now().Format(time.Stamp))
			// 	}
			// }
			ModdedNode = params.Hansen33Mod

			// if !ModdedNode {
			// 	fmt.Printf("%v Official Mining Node Detected - Happy Mining\n", time.Now().Format(time.Stamp))
			// }
			if config.Minimal {
				if params.Height != last_height || params.Difficultyuint64 != last_diff {
					last_height = params.Height
					last_diff = params.Difficultyuint64
					go SendTemplateToNodes(recv_data)
				}
			} else {
				go SendTemplateToNodes(recv_data)
			}
		}
	}
}

func SendUpdateToDaemon() {

	var count = 0
	for {
		if ModdedNode {
			if count == 0 {
				time.Sleep(60 * time.Second)
			}

			connection.WriteJSON(MinerInfo_Params{Wallet_Address: Address, Miner_Tag: "", Miner_Hashrate: Hashrate})

			count++
		}
		time.Sleep(10 * time.Second)
	}
}

func SendToDaemon(buffer []byte) {
	connection.WriteMessage(websocket.TextMessage, buffer)
}
