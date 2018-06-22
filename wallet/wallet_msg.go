package wallet

import (
	"gitlab.33.cn/chain33/chain33/types"
)

func (wallet *Wallet) ProcRecvMsg() {
	defer wallet.wg.Done()
	for msg := range wallet.client.Recv() {
		walletlog.Debug("wallet recv", "msg", msg.Id)
		msgtype := msg.Ty
		switch msgtype {

		case types.EventWalletGetAccountList:
			WalletAccounts, err := wallet.ProcGetAccountList()
			if err != nil {
				walletlog.Error("ProcGetAccountList", "err", err.Error())
				msg.Reply(wallet.client.NewMessage("rpc", types.EventWalletAccountList, err))
			} else {
				walletlog.Debug("process WalletAccounts OK")
				msg.Reply(wallet.client.NewMessage("rpc", types.EventWalletAccountList, WalletAccounts))
			}

		case types.EventWalletAutoMiner:
			flag := msg.GetData().(*types.MinerFlag).Flag
			if flag == 1 {
				wallet.walletStore.db.Set([]byte("WalletAutoMiner"), []byte("1"))
			} else {
				wallet.walletStore.db.Set([]byte("WalletAutoMiner"), []byte("0"))
			}
			wallet.setAutoMining(flag)
			wallet.flushTicket()
			msg.ReplyErr("WalletSetAutoMiner", nil)

		case types.EventWalletGetTickets:
			tickets, privs, err := wallet.GetTickets(1)
			if err != nil {
				walletlog.Error("GetTickets", "err", err.Error())
				msg.Reply(wallet.client.NewMessage("consensus", types.EventWalletTickets, err))
			} else {
				tks := &types.ReplyWalletTickets{tickets, privs}
				walletlog.Debug("process GetTickets OK")
				msg.Reply(wallet.client.NewMessage("consensus", types.EventWalletTickets, tks))
			}

		case types.EventNewAccount:
			NewAccount := msg.Data.(*types.ReqNewAccount)
			WalletAccount, err := wallet.ProcCreateNewAccount(NewAccount)
			if err != nil {
				walletlog.Error("ProcCreateNewAccount", "err", err.Error())
				msg.Reply(wallet.client.NewMessage("rpc", types.EventWalletAccount, err))
			} else {
				msg.Reply(wallet.client.NewMessage("rpc", types.EventWalletAccount, WalletAccount))
			}

		case types.EventWalletTransactionList:
			WalletTxList := msg.Data.(*types.ReqWalletTransactionList)
			TransactionDetails, err := wallet.ProcWalletTxList(WalletTxList)
			if err != nil {
				walletlog.Error("ProcWalletTxList", "err", err.Error())
				msg.Reply(wallet.client.NewMessage("rpc", types.EventTransactionDetails, err))
			} else {
				msg.Reply(wallet.client.NewMessage("rpc", types.EventTransactionDetails, TransactionDetails))
			}

		case types.EventWalletImportprivkey:
			ImportPrivKey := msg.Data.(*types.ReqWalletImportPrivKey)
			WalletAccount, err := wallet.ProcImportPrivKey(ImportPrivKey)
			if err != nil {
				walletlog.Error("ProcImportPrivKey", "err", err.Error())
				msg.Reply(wallet.client.NewMessage("rpc", types.EventWalletAccount, err))
			} else {
				msg.Reply(wallet.client.NewMessage("rpc", types.EventWalletAccount, WalletAccount))
			}
			wallet.flushTicket()

		case types.EventWalletSendToAddress:
			SendToAddress := msg.Data.(*types.ReqWalletSendToAddress)
			ReplyHash, err := wallet.ProcSendToAddress(SendToAddress)
			if err != nil {
				walletlog.Error("ProcSendToAddress", "err", err.Error())
				msg.Reply(wallet.client.NewMessage("rpc", types.EventReplyHashes, err))
			} else {
				msg.Reply(wallet.client.NewMessage("rpc", types.EventReplyHashes, ReplyHash))
			}

		case types.EventWalletSetFee:
			WalletSetFee := msg.Data.(*types.ReqWalletSetFee)

			var reply types.Reply
			reply.IsOk = true
			err := wallet.ProcWalletSetFee(WalletSetFee)
			if err != nil {
				walletlog.Error("ProcWalletSetFee", "err", err.Error())
				reply.IsOk = false
				reply.Msg = []byte(err.Error())
			}
			msg.Reply(wallet.client.NewMessage("rpc", types.EventReply, &reply))

		case types.EventWalletSetLabel:
			WalletSetLabel := msg.Data.(*types.ReqWalletSetLabel)
			WalletAccount, err := wallet.ProcWalletSetLabel(WalletSetLabel)

			if err != nil {
				walletlog.Error("ProcWalletSetLabel", "err", err.Error())
				msg.Reply(wallet.client.NewMessage("rpc", types.EventWalletAccount, err))
			} else {
				msg.Reply(wallet.client.NewMessage("rpc", types.EventWalletAccount, WalletAccount))
			}

		case types.EventWalletMergeBalance:
			MergeBalance := msg.Data.(*types.ReqWalletMergeBalance)
			ReplyHashes, err := wallet.ProcMergeBalance(MergeBalance)
			if err != nil {
				walletlog.Error("ProcMergeBalance", "err", err.Error())
				msg.Reply(wallet.client.NewMessage("rpc", types.EventReplyHashes, err))
			} else {
				msg.Reply(wallet.client.NewMessage("rpc", types.EventReplyHashes, ReplyHashes))
			}

		case types.EventWalletSetPasswd:
			SetPasswd := msg.Data.(*types.ReqWalletSetPasswd)

			var reply types.Reply
			reply.IsOk = true
			err := wallet.ProcWalletSetPasswd(SetPasswd)
			if err != nil {
				walletlog.Error("ProcWalletSetPasswd", "err", err.Error())
				reply.IsOk = false
				reply.Msg = []byte(err.Error())
			}
			msg.Reply(wallet.client.NewMessage("rpc", types.EventReply, &reply))

		case types.EventWalletLock:
			var reply types.Reply
			reply.IsOk = true
			err := wallet.ProcWalletLock()
			if err != nil {
				walletlog.Error("ProcWalletLock", "err", err.Error())
				reply.IsOk = false
				reply.Msg = []byte(err.Error())
			}
			msg.Reply(wallet.client.NewMessage("rpc", types.EventReply, &reply))

		case types.EventWalletUnLock:
			WalletUnLock := msg.Data.(*types.WalletUnLock)
			var reply types.Reply
			reply.IsOk = true
			err := wallet.ProcWalletUnLock(WalletUnLock)
			if err != nil {
				walletlog.Error("ProcWalletUnLock", "err", err.Error())
				reply.IsOk = false
				reply.Msg = []byte(err.Error())
			}
			msg.Reply(wallet.client.NewMessage("rpc", types.EventReply, &reply))
			wallet.flushTicket()

		case types.EventAddBlock:
			block := msg.Data.(*types.BlockDetail)
			wallet.ProcWalletAddBlock(block)
			walletlog.Debug("wallet add block --->", "height", block.Block.GetHeight())

		case types.EventDelBlock:
			block := msg.Data.(*types.BlockDetail)
			wallet.ProcWalletDelBlock(block)
			walletlog.Debug("wallet del block --->", "height", block.Block.GetHeight())

		//seed
		case types.EventGenSeed:
			genSeedLang := msg.Data.(*types.GenSeedLang)
			replySeed, err := wallet.genSeed(genSeedLang.Lang)
			if err != nil {
				walletlog.Error("genSeed", "err", err.Error())
				msg.Reply(wallet.client.NewMessage("rpc", types.EventReplyGenSeed, err))
			} else {
				msg.Reply(wallet.client.NewMessage("rpc", types.EventReplyGenSeed, replySeed))
			}

		case types.EventGetSeed:
			Pw := msg.Data.(*types.GetSeedByPw)
			seed, err := wallet.getSeed(Pw.Passwd)
			if err != nil {
				walletlog.Error("getSeed", "err", err.Error())
				msg.Reply(wallet.client.NewMessage("rpc", types.EventReplyGetSeed, err))
			} else {
				var replySeed types.ReplySeed
				replySeed.Seed = seed
				//walletlog.Error("EventGetSeed", "seed", seed)
				msg.Reply(wallet.client.NewMessage("rpc", types.EventReplyGetSeed, &replySeed))
			}

		case types.EventSaveSeed:
			saveseed := msg.Data.(*types.SaveSeedByPw)
			var reply types.Reply
			reply.IsOk = true
			ok, err := wallet.saveSeed(saveseed.Passwd, saveseed.Seed)
			if !ok {
				walletlog.Error("saveSeed", "err", err.Error())
				reply.IsOk = false
				reply.Msg = []byte(err.Error())
			}
			msg.Reply(wallet.client.NewMessage("rpc", types.EventReply, &reply))

		case types.EventGetWalletStatus:
			s := wallet.GetWalletStatus()
			msg.Reply(wallet.client.NewMessage("rpc", types.EventReplyWalletStatus, s))

		case types.EventDumpPrivkey:
			addr := msg.Data.(*types.ReqStr)
			privkey, err := wallet.ProcDumpPrivkey(addr.ReqStr)
			if err != nil {
				walletlog.Error("ProcDumpPrivkey", "err", err.Error())
				msg.Reply(wallet.client.NewMessage("rpc", types.EventReplyPrivkey, err))
			} else {
				var replyStr types.ReplyStr
				replyStr.Replystr = privkey
				msg.Reply(wallet.client.NewMessage("rpc", types.EventReplyPrivkey, &replyStr))
			}

		case types.EventCloseTickets:
			hashes, err := wallet.forceCloseTicket(wallet.GetHeight() + 1)
			if err != nil {
				walletlog.Error("closeTicket", "err", err.Error())
				msg.Reply(wallet.client.NewMessage("rpc", types.EventReplyHashes, err))
			} else {
				msg.Reply(wallet.client.NewMessage("rpc", types.EventReplyHashes, hashes))
				go func() {
					if len(hashes.Hashes) > 0 {
						wallet.waitTxs(hashes.Hashes)
						wallet.flushTicket()
					}
				}()
			}

		case types.EventSignRawTx:
			unsigned := msg.GetData().(*types.ReqSignRawTx)
			txHex, err := wallet.ProcSignRawTx(unsigned)
			if err != nil {
				walletlog.Error("EventSignRawTx", "err", err)
				msg.Reply(wallet.client.NewMessage("rpc", types.EventReplySignRawTx, err))
			} else {
				walletlog.Info("Reply EventSignRawTx", "msg", msg)
				msg.Reply(wallet.client.NewMessage("rpc", types.EventReplySignRawTx, &types.ReplySignRawTx{TxHex: txHex}))
			}
		case types.EventErrToFront: //收到系统发生致命性错误事件
			reportErrEvent := msg.Data.(*types.ReportErrEvent)
			wallet.setFatalFailure(reportErrEvent)
			walletlog.Debug("EventErrToFront")

		case types.EventFatalFailure: //定时查询是否有致命性故障产生
			fatalFailure := wallet.getFatalFailure()
			msg.Reply(wallet.client.NewMessage("rpc", types.EventReplyFatalFailure, &types.Int32{Data: fatalFailure}))

		default:
			walletlog.Info("ProcRecvMsg unknow msg", "msgtype", msgtype)
		}
		walletlog.Debug("end process", "msg.id", msg.Id)
	}
}