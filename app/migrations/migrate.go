package migrations

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/cosmos/cosmos-sdk/x/authz"

	tmjson "github.com/cometbft/cometbft/libs/json"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	ibcxfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibccoretypes "github.com/cosmos/ibc-go/v8/modules/core/types"

	evtypes "cosmossdk.io/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
	captypes "github.com/cosmos/ibc-go/modules/capability/types"

	legacy170 "github.com/scrtlabs/SecretNetwork/app/migrations/v170"
)

const (
	flagGenesisTime   = "genesis-time"
	flagInitialHeight = "initial-height"
)

// MigrateGenesisCmd returns a command to execute genesis state migration.
func MigrateGenesisCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate [genesis-file]",
		Short: "Migrate genesis to a specified target version",
		Long: `Migrate the source genesis into the target version and print to STDOUT.

Example:
$ secretd migrate /path/to/genesis.json --chain-id=secret-4 --genesis-time=2019-04-22T17:00:00Z --initial-height=5000
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			var err error

			importGenesis := args[0]

			jsonBlob, err := os.ReadFile(importGenesis)
			if err != nil {
				return errors.Wrap(err, "failed to read provided genesis file")
			}

			genDoc, err := tmtypes.GenesisDocFromJSON(jsonBlob)
			if err != nil {
				return errors.Wrapf(err, "failed to read genesis document from file %s", importGenesis)
			}

			// increase block consensus params
			genDoc.ConsensusParams.Block.MaxBytes = int64(10_000_000)
			genDoc.ConsensusParams.Block.MaxGas = int64(10_000_000)

			// decrease evidence max bytes
			genDoc.ConsensusParams.Evidence.MaxBytes = int64(50000)

			var initialState types.AppMap
			if err := json.Unmarshal(genDoc.AppState, &initialState); err != nil {
				return errors.Wrap(err, "failed to JSON unmarshal initial genesis state")
			}

			// Migrate 120 state to 170 state
			newGenState := legacy170.Migrate(initialState, clientCtx)

			var bankGenesis banktypes.GenesisState

			clientCtx.Codec.MustUnmarshalJSON(newGenState[banktypes.ModuleName], &bankGenesis)

			var stakingGenesis staking.GenesisState

			clientCtx.Codec.MustUnmarshalJSON(newGenState[staking.ModuleName], &stakingGenesis)

			ibcTransferGenesis := ibcxfertypes.DefaultGenesisState()
			ibcCoreGenesis := ibccoretypes.DefaultGenesisState()
			capGenesis := captypes.DefaultGenesis()
			evGenesis := evtypes.DefaultGenesisState()
			authzGenesis := authz.DefaultGenesisState()

			ibcTransferGenesis.Params.ReceiveEnabled = false
			ibcTransferGenesis.Params.SendEnabled = false

			ibcCoreGenesis.ClientGenesis.Params.AllowedClients = []string{exported.Tendermint}
			stakingGenesis.Params.HistoricalEntries = 10000

			newGenState[ibcxfertypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(ibcTransferGenesis)
			newGenState[host.SubModuleName] = clientCtx.Codec.MustMarshalJSON(ibcCoreGenesis)
			newGenState[captypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(capGenesis)
			newGenState[evtypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(evGenesis)
			newGenState[staking.ModuleName] = clientCtx.Codec.MustMarshalJSON(&stakingGenesis)
			newGenState[authz.ModuleName] = clientCtx.Codec.MustMarshalJSON(authzGenesis)

			genDoc.AppState, err = json.Marshal(newGenState)
			if err != nil {
				return errors.Wrap(err, "failed to JSON marshal migrated genesis state")
			}

			genesisTime, _ := cmd.Flags().GetString(flagGenesisTime)
			if genesisTime != "" {
				var t time.Time

				err := t.UnmarshalText([]byte(genesisTime))
				if err != nil {
					return errors.Wrap(err, "failed to unmarshal genesis time")
				}

				genDoc.GenesisTime = t
			}

			chainID, _ := cmd.Flags().GetString(flags.FlagChainID)
			if chainID != "" {
				genDoc.ChainID = chainID
			}

			initialHeight, _ := cmd.Flags().GetInt(flagInitialHeight)

			genDoc.InitialHeight = int64(initialHeight)

			bz, err := tmjson.Marshal(genDoc)
			if err != nil {
				return errors.Wrap(err, "failed to marshal genesis doc")
			}

			sortedBz, err := sdk.SortJSON(bz)
			if err != nil {
				return errors.Wrap(err, "failed to sort JSON genesis doc")
			}

			fmt.Println(string(sortedBz))
			return nil
		},
	}

	cmd.Flags().String(flagGenesisTime, "", "override genesis_time with this flag")
	cmd.Flags().Int(flagInitialHeight, 0, "Set the starting height for the chain")
	cmd.Flags().String(flags.FlagChainID, "", "override chain_id with this flag")

	return cmd
}
