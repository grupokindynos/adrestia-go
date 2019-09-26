package bitshares

type CBBalance struct {
	Data []CBAsset `json:"data"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

type CBAsset struct {
	Amount float64 `json:"amount"`
	Asset  struct {
		// Description        interface{} `json:"description"`
		DynamicAssetDataID string `json:"dynamic_asset_data_id"`
		Flags              struct {
			ChargeMarketFee     bool `json:"charge_market_fee"`
			CommitteeFedAsset   bool `json:"committee_fed_asset"`
			DisableConfidential bool `json:"disable_confidential"`
			DisableForceSettle  bool `json:"disable_force_settle"`
			GlobalSettle        bool `json:"global_settle"`
			OverrideAuthority   bool `json:"override_authority"`
			TransferRestricted  bool `json:"transfer_restricted"`
			WhiteList           bool `json:"white_list"`
			WitnessFedAsset     bool `json:"witness_fed_asset"`
		} `json:"flags"`
		ID      string `json:"id"`
		Issuer  string `json:"issuer"`
		Options struct {
			BlacklistAuthorities []string `json:"blacklist_authorities"`
			BlacklistMarkets     []string `json:"blacklist_markets"`
			CoreExchangeRate     struct {
				Base struct {
					Amount  int    `json:"amount"`
					AssetID string `json:"asset_id"`
				} `json:"base"`
				Quote struct {
					Amount  int    `json:"amount"`
					AssetID string `json:"asset_id"`
				} `json:"quote"`
			} `json:"core_exchange_rate"`
			Description string `json:"description"`
			Extensions  struct {
			} `json:"extensions"`
			Flags                int           `json:"flags"`
			IssuerPermissions    int           `json:"issuer_permissions"`
			MarketFeePercent     int           `json:"market_fee_percent"`
			MaxMarketFee         string        `json:"max_market_fee"`
			MaxSupply            string        `json:"max_supply"`
			WhitelistAuthorities []string `json:"whitelist_authorities"`
			WhitelistMarkets     []string `json:"whitelist_markets"`
		} `json:"options"`
		Permissions struct {
			ChargeMarketFee     bool `json:"charge_market_fee"`
			CommitteeFedAsset   bool `json:"committee_fed_asset"`
			DisableConfidential bool `json:"disable_confidential"`
			DisableForceSettle  bool `json:"disable_force_settle"`
			GlobalSettle        bool `json:"global_settle"`
			OverrideAuthority   bool `json:"override_authority"`
			TransferRestricted  bool `json:"transfer_restricted"`
			WhiteList           bool `json:"white_list"`
			WitnessFedAsset     bool `json:"witness_fed_asset"`
		} `json:"permissions"`
		Precision int    `json:"precision"`
		Symbol    string `json:"symbol"`
	} `json:"asset"`
	Symbol string `json:"symbol"`
}

type CBRates []struct {
	ID            string  `json:"id"`
	Last          string  `json:"last"`
	Volume        string  `json:"volume"`
	Ask           string  `json:"ask"`
	Bid           string  `json:"bid"`
	PercentChange float64 `json:"percentChange"`
	Restricted    bool    `json:"restricted"`
}

type Orders struct {
	Data struct {
		Asks []TradeObject `json:"asks"`
		Bids []TradeObject `json:"bids"`
	} `json:"data"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

// This object can either be an ask or a Bid
type TradeObject struct {
	Base struct {
		Amount float64 `json:"amount"`
		Asset  struct {
			Description struct {
				Main      string `json:"main"`
				Market    string `json:"market"`
				ShortName string `json:"short_name"`
			} `json:"description"`
			DynamicAssetDataID string `json:"dynamic_asset_data_id"`
			Flags              struct {
				ChargeMarketFee     bool `json:"charge_market_fee"`
				CommitteeFedAsset   bool `json:"committee_fed_asset"`
				DisableConfidential bool `json:"disable_confidential"`
				DisableForceSettle  bool `json:"disable_force_settle"`
				GlobalSettle        bool `json:"global_settle"`
				OverrideAuthority   bool `json:"override_authority"`
				TransferRestricted  bool `json:"transfer_restricted"`
				WhiteList           bool `json:"white_list"`
				WitnessFedAsset     bool `json:"witness_fed_asset"`
			} `json:"flags"`
			ID      string `json:"id"`
			Issuer  string `json:"issuer"`
			Options struct {
				BlacklistAuthorities []interface{} `json:"blacklist_authorities"`
				BlacklistMarkets     []interface{} `json:"blacklist_markets"`
				CoreExchangeRate     struct {
					Base struct {
						Amount  int    `json:"amount"`
						AssetID string `json:"asset_id"`
					} `json:"base"`
					Quote struct {
						Amount  int    `json:"amount"`
						AssetID string `json:"asset_id"`
					} `json:"quote"`
				} `json:"core_exchange_rate"`
				Description string `json:"description"`
				Extensions  struct {
				} `json:"extensions"`
				Flags                int           `json:"flags"`
				IssuerPermissions    int           `json:"issuer_permissions"`
				MarketFeePercent     int           `json:"market_fee_percent"`
				MaxMarketFee         string        `json:"max_market_fee"`
				MaxSupply            string        `json:"max_supply"`
				WhitelistAuthorities []interface{} `json:"whitelist_authorities"`
				WhitelistMarkets     []interface{} `json:"whitelist_markets"`
			} `json:"options"`
			Permissions struct {
				ChargeMarketFee     bool `json:"charge_market_fee"`
				CommitteeFedAsset   bool `json:"committee_fed_asset"`
				DisableConfidential bool `json:"disable_confidential"`
				DisableForceSettle  bool `json:"disable_force_settle"`
				GlobalSettle        bool `json:"global_settle"`
				OverrideAuthority   bool `json:"override_authority"`
				TransferRestricted  bool `json:"transfer_restricted"`
				WhiteList           bool `json:"white_list"`
				WitnessFedAsset     bool `json:"witness_fed_asset"`
			} `json:"permissions"`
			Precision int    `json:"precision"`
			Symbol    string `json:"symbol"`
		} `json:"asset"`
		Symbol string `json:"symbol"`
	} `json:"base"`
}
