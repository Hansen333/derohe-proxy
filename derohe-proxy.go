package main

import (
	"derohe-proxy/config"
	"derohe-proxy/proxy"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/docopt/docopt-go"
)

func main() {
	var err error
	// var rwmutex sync.RWMutex

	config.Arguments, err = docopt.Parse(config.Command_line, nil, true, "pre-alpha", false)

	if err != nil {
		return
	}

	if config.Arguments["--listen-address"] != nil {
		addr, err := net.ResolveTCPAddr("tcp", config.Arguments["--listen-address"].(string))
		if err != nil {
			return
		} else {
			if addr.Port == 0 {
				return
			} else {
				config.Listen_addr = addr.String()
			}
		}
	}

	if config.Arguments["--daemon-address"] == nil {
		return
	} else {
		config.Daemon_address = config.Arguments["--daemon-address"].(string)
	}

	if config.Arguments["--log-interval"] != nil {
		interval, err := strconv.ParseInt(config.Arguments["--log-interval"].(string), 10, 32)
		if err != nil {
			return
		} else {
			if interval < 1 || interval > 3600 {
				config.Log_intervall = 60
			} else {
				config.Log_intervall = int(interval)
			}
		}
	}

	if config.Arguments["--minimal"].(bool) {
		config.Minimal = true
		fmt.Printf("%v Forward only 2 jobs per block\n", time.Now().Format(time.Stamp))
	}

	if config.Arguments["--nonce"].(bool) {
		config.Nonce = true
		fmt.Printf("%v Nonce editing is enabled\n", time.Now().Format(time.Stamp))
	}

	if config.Arguments["--pool"].(bool) {
		config.Pool_mode = true
		config.Minimal = false
		fmt.Printf("%v Pool mode is enabled\n", time.Now().Format(time.Stamp))
	}

	fmt.Printf("%v Logging every %d seconds\n", time.Now().Format(time.Stamp), config.Log_intervall)

	go proxy.Start_server()

	// Wait for first miner connection to grab wallet address
	for proxy.CountMiners() < 1 {
		time.Sleep(time.Second * 1)
	}
	go proxy.Start_client(proxy.Address)
	go proxy.SendUpdateToDaemon()

	fmt.Print("M == Miners - V == Velocity (T per 10 min) - T == Total Blocks Shares - S == Session - OL == Orphan Loss\n")
	for {
		time.Sleep(time.Second * time.Duration(config.Log_intervall))

		hash_rate_string := ""

		switch {
		case proxy.Hashrate > 1000000000000:
			hash_rate_string = fmt.Sprintf("%.3f TH/s", float64(proxy.Hashrate)/1000000000000.0)
		case proxy.Hashrate > 1000000000:
			hash_rate_string = fmt.Sprintf("%.3f GH/s", float64(proxy.Hashrate)/1000000000.0)
		case proxy.Hashrate > 1000000:
			hash_rate_string = fmt.Sprintf("%.3f MH/s", float64(proxy.Hashrate)/1000000.0)
		case proxy.Hashrate > 1000:
			hash_rate_string = fmt.Sprintf("%.3f KH/s", float64(proxy.Hashrate)/1000.0)
		case proxy.Hashrate > 0:
			hash_rate_string = fmt.Sprintf("%d H/s", int(proxy.Hashrate))
		}

		Velocity := float64(0)

		total_blocks := proxy.Blocks + proxy.Minis

		if total_blocks >= 1 {
			Velocity = (float64(total_blocks) / ((float64(time.Now().Unix()) - float64(proxy.ProxyStart.Unix())) / 600))
		}

		orphan_loss := float64(0)

		if proxy.Orphans >= 1 && total_blocks >= 1 {
			orphan_loss = float64(float64(float64(float64(proxy.Orphans)/float64(total_blocks))) * 100)
		}

		// l.SetPrompt(fmt.Sprintf("\033[1m\033[32mDERO HE (\033[31m%s-mod\033[32m):%s \033[0m"+color+"%d/%d [%d/%d] "+pcolor+"P %d/%d TXp %d:%d \033[32mNW %s >MN %d/%d [%d/%d] %s>>\033[0m ",

		orphan_color := "\033[0m"
		if orphan_loss > 1.5 {
			orphan_color = "\033[31m"
		}

		fmt.Printf("\r[ %s ] M:\033[32m%d\033[0m V:\033[34m%.4f\033[0m T:%d OL:"+orphan_color+"%.3f\033[0m S:%d IB:%d MB:%d MBR:%d MBO:%d MINING @ \033[32m%s\033[0m ...", time.Now().Sub(proxy.ProxyStart).Round(time.Second).String(), proxy.CountMiners(), Velocity, proxy.Shares, orphan_loss, proxy.ReconnectCount, proxy.Blocks, proxy.Minis, proxy.Rejected, proxy.Orphans, hash_rate_string)

		// rwmutex.RLock()
		// // for i := range proxy.Wallet_count {
		// // 	if proxy.Wallet_count[i] > 1 {
		// // 		fmt.Printf("%v Wallet %v, %d miners\n", time.Now().Format(time.Stamp), i, proxy.Wallet_count[i])
		// // 	}
		// // }
		// rwmutex.RUnlock()
	}
}
