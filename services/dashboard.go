/*---------------------------------------------------------------------------------------------
 *  Copyright (c) IBAX All rights reserved.
 *  See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/
package services

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/IBAX-io/go-explorer/models"
	"github.com/sirupsen/logrus"
)

var (
	GDashboardDBTransactions []models.DBTransactionsInfo
	GDashboardChain          map[string]interface{}
	GMaxblockid              int64
	SendWebsocketData        chan bool
)

func Deal_Redis_Dashboard() error {
	var (
		err error
	)

	bk := &models.Block{}
	ret, err := bk.GetMaxBlock()
	if err == nil && ret {
		//if GMaxblockid != bk.ID {
		GMaxblockid = bk.ID
		_, err = Deal_Dashboard_top_num()
		if err != nil {
			logrus.Info("Deal_Dashboard_top_num" + err.Error())
			return err
		}
		//}
	}
	return err
}

func Get_history_map() (*[]models.DBTransactionsInfo, error) {
	var (
		err error
	)
	trans, err := models.GetDBDayTraninfo(30)
	if err != nil {
		return trans, err
	} else {
		GDashboardDBTransactions = *trans
		Set_Redis_Dashboard_history_map(trans)
	}
	return trans, err

}

func DealRedisDashboardGetChain() (*map[string]interface{}, error) {
	var (
		err error
	)
	if GDashboardChain == nil {
		GDashboardChain = make(map[string]interface{})
	}
	trans, err := Getchain()
	if err != nil {
		return trans, err
	} else {
		Set_Redis_Dashboard_Get_chain(trans)
	}
	return trans, err

}

func Set_Redis_Dashboard_history_map(dat *[]models.DBTransactionsInfo) error {
	lg1, err := json.Marshal(dat)
	if err != nil {
		return err
	}
	rp := models.RedisParams{
		Key:   "Dashboard_history_map",
		Value: string(lg1),
	}
	err = rp.Set()
	if err != nil {
		return err
	}
	return nil

}

func Set_Redis_Dashboard_Get_chain(dat *map[string]interface{}) error {
	lg1, err := json.Marshal(dat)
	if err != nil {
		return err
	}
	rp := &models.RedisParams{
		Key:   "Dashboard_Get_ibax",
		Value: string(lg1),
	}
	err = rp.Set()
	if err != nil {
		return err
	}
	return nil
}

func Deal_Dashboard_top_num() (*models.ResponseTopTitle, error) {
	dat := models.ResponseTopTitle{}
	dat.Cmd = models.CmdTopBlocks
	topblocks, err := Get_Group_Block_Lists(1, 10, "time desc")
	if err != nil {
		return &dat, err
	}
	dat.Data = topblocks
	if err := SendChannelTopData(dat); err != nil {
		logrus.Info("Send top transaction err:" + err.Error())
	}

	//das := *topblocks
	//blockid := das[0].BlockID
	//old example:time:1629786700 tx:8
	//topblocktps, err := Get_Group_Block_TpsLists(1, 15, "time desc")
	//if err != nil {
	//	return &dat, err
	//}
	//dat.Data = topblocktps
	//dat.Cmd = models.CmdTopTransactiontps
	//if err := SendChannelDashboard(dat); err != nil {
	//	return &dat, err
	//}

	toptransactions, err := Get_Group_TransactionBlockLast()
	if err != nil {
		return &dat, err
	}
	dat.Data = toptransactions
	dat.Cmd = models.CmdTopTransactions
	if err := SendChannelTopData(dat); err != nil {
		logrus.Info("Send top transaction err:" + err.Error())
		return &dat, err
	}

	return &dat, nil
}
func SendChannelTopData(dat models.ResponseTopTitle) error {
	ds, err := json.Marshal(dat)
	if err != nil {
		return err
	}
	err = WriteChannelByte(models.ChannelTopData, ds)
	if err != nil {
		return err
	}
	return nil
}
func Get_Dashboard_history_map() (*[]models.DBTransactionsInfo, error) {
	if GDashboardDBTransactions != nil {
		return &GDashboardDBTransactions, nil
	}
	ret, err := Get_history_map()
	return ret, err
}

func Get_Dashboard_Get_chain() (*map[string]interface{}, error) {
	if GDashboardChain != nil {
		GetCheckchain()
		return &GDashboardChain, nil
	}
	ret, err := DealRedisDashboardGetChain()
	return ret, err
}

func GetCheckchain() {
	if GDashboardChain != nil {
		if _, ok := GDashboardChain["btc-ibax"]; ok {
		} else {
			if ret1, err1 := Getchaindat("https://api.coinegg.im/api/v1/ticker/region/btc?coin=ibxc"); err1 == nil {
				GDashboardChain["btc-ibax"] = ret1
			} else {
				logrus.Info("btc-ibax Not Found")
			}
		}

		if _, ok := GDashboardChain["usdt-ibax"]; ok {
		} else {
			if ret2, err2 := Getchaindat("https://api.coinegg.im/api/v1/ticker/region/usdt?coin=ibxc"); err2 == nil {
				GDashboardChain["usdt-ibax"] = ret2
			} else {
				logrus.Info("usdt-ibax Not Found")
			}
		}

		if _, ok := GDashboardChain["Rates"]; ok {
		} else {
			if ret3, err3 := GetBlockCCRatesdat("https://data.block.cc/api/v1/exchange_rate?base=usdt"); err3 == nil {
				GDashboardChain["Rates"] = ret3
			} else {
				logrus.Info("Rates Not Found")
			}
		}

		if _, ok := GDashboardChain["Price"]; ok {
		} else {
			if ret4, err4 := GetBlockCCPricedat("https://data.block.cc/api/v1/price?symbol_name=bitcoin"); err4 == nil {
				GDashboardChain["Price"] = ret4
			} else {
				logrus.Info("Price Not Found")
			}

		}

	}

	//return  err
}

func Getchain() (*map[string]interface{}, error) {
	var (
		err error
	)
	rets := make(map[string]interface{})
	if ret1, err1 := Getchaindat("https://api.coinegg.im/api/v1/ticker/region/btc?coin=ibxc"); err1 == nil {
		//rets = append(rets, *ret1)
		rets["btc-ibax"] = ret1
		GDashboardChain["btc-ibax"] = ret1
	} else {
		if GDashboardChain != nil {
			if v, ok := GDashboardChain["btc-ibax"]; ok {
				//fmt.Println(v)
				rets["btc-ibax"] = v
			} else {
				logrus.Info("btc-ibax Not Found")
				//fmt.Println("Key Not Found")
			}

		}
		err = err1
	}

	if ret2, err2 := Getchaindat("https://api.coinegg.im/api/v1/ticker/region/usdt?coin=ibxc"); err2 == nil {
		//rets = append(rets, *ret2)
		rets["usdt-ibax"] = ret2
		GDashboardChain["usdt-ibax"] = ret2
	} else {
		if GDashboardChain != nil {
			if v, ok := GDashboardChain["usdt-ibax"]; ok {
				//fmt.Println(v)
				rets["usdt-ibax"] = v
			} else {
				logrus.Info("usdt-ibax Not Found")
				//fmt.Println("Key Not Found")
			}

		}
		err = err2
	}

	if ret3, err3 := GetBlockCCRatesdat("https://data.block.cc/api/v1/exchange_rate?base=usdt"); err3 == nil {
		//rets = append(rets, *ret1)
		rets["Rates"] = ret3
		GDashboardChain["Rates"] = ret3
	} else {
		if GDashboardChain != nil {
			if v, ok := GDashboardChain["Rates"]; ok {
				//fmt.Println(v)
				rets["Rates"] = v
			} else {
				logrus.Info("Rates Not Found")
				//fmt.Println("Key Not Found")
			}

		}
		err = err3
	}

	if ret4, err4 := GetBlockCCPricedat("https://data.block.cc/api/v1/price?symbol_name=bitcoin"); err4 == nil {
		//rets = append(rets, *ret2)
		rets["Price"] = ret4
		GDashboardChain["Price"] = ret4
	} else {
		if GDashboardChain != nil {
			if v, ok := GDashboardChain["Price"]; ok {
				//fmt.Println(v)
				rets["Price"] = v
			} else {
				logrus.Info("Price Not Found")
				//fmt.Println("Key Not Found")
			}

		}
		err = err4
	}

	return &rets, err
}
func Getchaindat(path string) (*models.DashboardChainInfo, error) {
	ret := &models.DashboardChainInfo{}
	client := &http.Client{}
	request, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	request.Header.Set("Accept-Charset", "zh-CN,zh;q=0.9")
	request.Header.Set("Accept-Encoding", "")
	request.Header.Set("Accept-Language", "zh-CN,zh;q=0.8")
	request.Header.Set("Cache-Control", "max-age=0")
	request.Header.Set("Connection", "keep-alive")

	if response, err := client.Do(request); err == nil {
		defer response.Body.Close()
		if response.StatusCode == 200 {
			if body, err := ioutil.ReadAll(response.Body); err == nil {
				if err := json.Unmarshal(body, ret); err == nil {
					return ret, err
				}
				return ret, err
			}
			return ret, err
		}
		return ret, err
	}

	return ret, err
}
func GetBlockCCRatesdat(path string) (*models.RatesInfo, error) {
	retdat := &models.BlockCCRatesInfo{}
	ret := &models.RatesInfo{}
	client := &http.Client{}
	//defer  client.
	request, err := http.NewRequest("GET", path, nil)
	request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	request.Header.Set("Accept-Charset", "zh-CN,zh;q=0.9")
	request.Header.Set("Accept-Encoding", "")
	request.Header.Set("Accept-Language", "zh-CN,zh;q=0.8")
	request.Header.Set("Cache-Control", "max-age=0")
	request.Header.Set("Connection", "keep-alive")

	if response, err := client.Do(request); err == nil {
		defer response.Body.Close()
		if response.StatusCode == 200 {
			if body, err := ioutil.ReadAll(response.Body); err == nil {
				if err := json.Unmarshal(body, retdat); err != nil {
					return &retdat.Data, err
				} else {
					return &retdat.Data, err
				}

				return ret, err
			}
			return ret, err
		}
		return ret, err
	}

	return ret, err
}
func GetBlockCCPricedat(path string) (*[]map[string]interface{}, error) {
	retdat := &models.BlockCCPriceInfo{}
	ret := make([]map[string]interface{}, 1, 1)
	//ret := &[]map[string]interface{}
	client := &http.Client{}
	request, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	request.Header.Set("Accept-Charset", "zh-CN,zh;q=0.9")
	request.Header.Set("Accept-Encoding", "")
	request.Header.Set("Accept-Language", "zh-CN,zh;q=0.8")
	request.Header.Set("Cache-Control", "max-age=0")
	request.Header.Set("Connection", "keep-alive")

	if response, err := client.Do(request); err == nil {
		defer response.Body.Close()
		if response.StatusCode == 200 {
			if body, err := ioutil.ReadAll(response.Body); err == nil {
				if err := json.Unmarshal(body, retdat); err != nil {
					return &retdat.Data, err
				} else {
					return &retdat.Data, err
				}
			}
			return &ret, err
		}
		return &ret, err
	}

	return &ret, err
}
