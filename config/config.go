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

package config

const (
	OntUsefulBlockNum = 1
)

type Config struct {
	ZilConfig       *ZILConfig
	PolyConfig      *POLYConfig
	Path            string
	RemoveDB        bool
	TargetContracts []map[string]map[string][]uint64
}

type ZILConfig struct {
	ZilApiEndpoint                 string
	ZilChainId                     int
	ZilMessageVersion              int
	ZilMonitorInterval             uint32
	ZilHeadersPerBatch             uint32
	ZilForceHeight                 uint64
	SideChainId                    uint64
	CrossChainManagerContract      string
	CrossChainManagerProxyContract string
	MaxExistTxEpoch                int
	KeyStorePath                   string
	KeyStorePwdSet                 map[string]interface{}
}

type POLYConfig struct {
	PolyWalletFile          string
	PolyWalletPassword      string
	PolyStartHeight         uint32
	PolyMonitorInterval     uint32
	EntranceContractAddress string
	RestUrl                 string
}
