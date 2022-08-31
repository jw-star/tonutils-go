package wallet

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

// https://tonscan.org/address/EQAaQOzG_vqjGo71ZJNiBdU1SRenbqhEzG8vfpZwubzyB0T8
const _V4R1CodeHex = "b5ee9c72410215010002f5000114ff00f4a413f4bcf2c80b010201200203020148040504f8f28308d71820d31fd31fd31f02f823bbf263ed44d0d31fd31fd3fff404d15143baf2a15151baf2a205f901541064f910f2a3f80024a4c8cb1f5240cb1f5230cbff5210f400c9ed54f80f01d30721c0009f6c519320d74a96d307d402fb00e830e021c001e30021c002e30001c0039130e30d03a4c8cb1f12cb1fcbff1112131403eed001d0d3030171b0915be021d749c120915be001d31f218210706c7567bd228210626c6e63bdb022821064737472bdb0925f03e002fa403020fa4401c8ca07cbffc9d0ed44d0810140d721f404305c810108f40a6fa131b3925f05e004d33fc8258210706c7567ba9131e30d248210626c6e63bae30004060708020120090a005001fa00f404308210706c7567831eb17080185005cb0527cf165003fa02f40012cb69cb1f5210cb3f0052f8276f228210626c6e63831eb17080185005cb0527cf1624fa0214cb6a13cb1f5230cb3f01fa02f4000092821064737472ba8e3504810108f45930ed44d0810140d720c801cf16f400c9ed54821064737472831eb17080185004cb0558cf1622fa0212cb6acb1fcb3f9410345f04e2c98040fb000201200b0c0059bd242b6f6a2684080a06b90fa0218470d4080847a4937d29910ce6903e9ff9837812801b7810148987159f31840201580d0e0011b8c97ed44d0d70b1f8003db29dfb513420405035c87d010c00b23281f2fff274006040423d029be84c600201200f100019adce76a26840206b90eb85ffc00019af1df6a26840106b90eb858fc0006ed207fa00d4d422f90005c8ca0715cbffc9d077748018c8cb05cb0222cf165005fa0214cb6b12ccccc971fb00c84014810108f451f2a702006c810108d718c8542025810108f451f2a782106e6f746570748018c8cb05cb025004cf16821005f5e100fa0213cb6a12cb1fc971fb00020072810108d718305202810108f459f2a7f82582106473747270748018c8cb05cb025005cf16821005f5e100fa0214cb6a13cb1f12cb3fc973fb00000af400c9ed5446a9f34f"

// https://github.com/toncenter/tonweb/blob/master/src/contract/wallet/WalletSources.md#revision-2-3
const _V4R2CodeHex = "B5EE9C72410214010002D4000114FF00F4A413F4BCF2C80B010201200203020148040504F8F28308D71820D31FD31FD31F02F823BBF264ED44D0D31FD31FD3FFF404D15143BAF2A15151BAF2A205F901541064F910F2A3F80024A4C8CB1F5240CB1F5230CBFF5210F400C9ED54F80F01D30721C0009F6C519320D74A96D307D402FB00E830E021C001E30021C002E30001C0039130E30D03A4C8CB1F12CB1FCBFF1011121302E6D001D0D3032171B0925F04E022D749C120925F04E002D31F218210706C7567BD22821064737472BDB0925F05E003FA403020FA4401C8CA07CBFFC9D0ED44D0810140D721F404305C810108F40A6FA131B3925F07E005D33FC8258210706C7567BA923830E30D03821064737472BA925F06E30D06070201200809007801FA00F40430F8276F2230500AA121BEF2E0508210706C7567831EB17080185004CB0526CF1658FA0219F400CB6917CB1F5260CB3F20C98040FB0006008A5004810108F45930ED44D0810140D720C801CF16F400C9ED540172B08E23821064737472831EB17080185005CB055003CF1623FA0213CB6ACB1FCB3FC98040FB00925F03E20201200A0B0059BD242B6F6A2684080A06B90FA0218470D4080847A4937D29910CE6903E9FF9837812801B7810148987159F31840201580C0D0011B8C97ED44D0D70B1F8003DB29DFB513420405035C87D010C00B23281F2FFF274006040423D029BE84C600201200E0F0019ADCE76A26840206B90EB85FFC00019AF1DF6A26840106B90EB858FC0006ED207FA00D4D422F90005C8CA0715CBFFC9D077748018C8CB05CB0222CF165005FA0214CB6B12CCCCC973FB00C84014810108F451F2A7020070810108D718FA00D33FC8542047810108F451F2A782106E6F746570748018C8CB05CB025006CF165004FA0214CB6A12CB1FCB3FC973FB0002006C810108D718FA00D33F305224810108F459F2A782106473747270748018C8CB05CB025005CF165003FA0213CB6ACB1F12CB3FC973FB00000AF400C9ED54696225E5"

var (
	_V4R1CodeBOC []byte
	_V4R2CodeBOC []byte
)

func init() {
	var err error

	_V4R1CodeBOC, err = hex.DecodeString(_V4R1CodeHex)
	if err != nil {
		panic(err)
	}
	_V4R2CodeBOC, err = hex.DecodeString(_V4R2CodeHex)
	if err != nil {
		panic(err)
	}
}

type SpecV4R2 struct {
	SpecRegular
}

func (s *SpecV4R2) BuildMessage(ctx context.Context, isInitialized bool, block *tlb.BlockInfo, messages []*Message) (*cell.Cell, error) {
	if len(messages) > 4 {
		return nil, errors.New("for this type of wallet max 4 messages can be sent in the same time")
	}

	var seq uint64

	if isInitialized {
		resp, err := s.wallet.api.RunGetMethod(ctx, block, s.wallet.addr, "seqno")
		if err != nil {
			return nil, fmt.Errorf("get seqno err: %w", err)
		}

		iSeq, ok := resp[0].(int64)
		if !ok {
			return nil, fmt.Errorf("seqno is not an integer")
		}
		seq = uint64(iSeq)
	}

	payload := cell.BeginCell().MustStoreUInt(uint64(s.wallet.subwallet), 32).
		MustStoreUInt(uint64(timeNow().Add(time.Duration(s.messagesTTL)*time.Second).UTC().Unix()), 32).
		MustStoreUInt(seq, 32).
		MustStoreInt(0, 8) // op

	for i, message := range messages {
		intMsg, err := message.InternalMessage.ToCell()
		if err != nil {
			return nil, fmt.Errorf("failed to convert internal message %d to cell: %w", i, err)
		}

		payload.MustStoreUInt(uint64(message.Mode), 8).MustStoreRef(intMsg)
	}

	sign := payload.EndCell().Sign(s.wallet.key)
	msg := cell.BeginCell().MustStoreSlice(sign, 512).MustStoreBuilder(payload).EndCell()

	return msg, nil
}

// TODO: implement plugins
