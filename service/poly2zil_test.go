/*
 * Copyright (C) 2021 Zilliqa
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package service

import (
	"fmt"
	"github.com/Zilliqa/gozilliqa-sdk/provider"
	"github.com/magiconair/properties/assert"
	"github.com/polynetwork/zilliqa-relayer/config"
	"testing"
)

var p *PolySyncManager
var s *ZilSender

func init() {
	zilSdk := provider.NewProvider("https://polynetworkcc3dcb2-5-api.dev.z7a.xyz")
	p = &PolySyncManager{
		zilSdk: zilSdk,
		cfg: &config.Config{ZilConfig: &config.ZILConfig{
			CrossChainManagerContract: "zil1ur4vwcmcz3jqypksgq7qeju2sk5jrskzaadau5",
		}},
	}
	s = &ZilSender{
		cfg: &config.Config{ZilConfig: &config.ZILConfig{
			CrossChainManagerContract: "zil1ur4vwcmcz3jqypksgq7qeju2sk5jrskzaadau5",
		}},
		zilSdk:  zilSdk,
		address: "zil1ur4vwcmcz3jqypksgq7qeju2sk5jrskzaadau5",
	}
}

func TestPolySyncManager_FindLatestHeight(t *testing.T) {
	fmt.Println(p.findLatestHeight())
}

func TestZilSender_CheckIfFromChainTxExist(t *testing.T) {
	exist := s.checkIfFromChainTxExist(3, "0x00ca93f8738111a063d8ab7221f47c70a4cade0ca4a2829df494cd4b5e231bd6")
	assert.Equal(t, exist, true)

	exist = s.checkIfFromChainTxExist(3, "0x00ca93f8738111a063d8ab7221f47c70a4cade0ca4a2829df494cd4b5e231bd7")
	assert.Equal(t, exist, false)

}
