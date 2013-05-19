package main

import (
	"fmt"
	"os"
	"encoding/hex"
	"github.com/piotrnar/gocoin/btc"
)

func load_tx(par string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Something went wrong, but recovered in f", r)
		}
	}()

	f, e := os.Open(par)
	if e != nil {
		println(e.Error())
		return
	}
	n, _ := f.Seek(0, os.SEEK_END)
	f.Seek(0, os.SEEK_SET)
	buf := make([]byte, n)
	f.Read(buf)
	f.Close()

	txd, er := hex.DecodeString(string(buf))
	if er != nil {
		txd = buf
		fmt.Println("Seems like the transaction is in a binary format")
	} else {
		fmt.Println("Looks like the transaction file contains hex data")
	}

	// At this place we should have raw transaction in txd
	tx, le := btc.NewTx(txd)
	if le != len(txd) {
		fmt.Println("WARNING: Tx length mismatch", le, len(txd))
	}
	txid := btc.NewSha2Hash(txd)
	fmt.Println(len(tx.TxIn), "Inputs:")
	var totinp, totout uint64
	var missinginp bool
	for i := range tx.TxIn {
		fmt.Printf(" %3d %s", i, tx.TxIn[i].Input.String())
		po, _ := BlockChain.Unspent.UnspentGet(&tx.TxIn[i].Input)
		if po != nil {
			ok := btc.VerifyTxScript(tx.TxIn[i].ScriptSig, po.Pk_script, i, tx)
			if !ok {
				fmt.Println("The transacion does not have a valid signature!")
				return
			}
			totinp += po.Value
			fmt.Printf(" %15.8f BTC @ %s\n", float64(po.Value)/1e8,
				btc.NewAddrFromPkScript(po.Pk_script, AddrVersion).String())
		} else {
			fmt.Println(" * no such unspent in the blockchain *")
			missinginp = true
		}
	}
	fmt.Println(len(tx.TxOut), "Outputs:")
	for i := range tx.TxOut {
		totout += tx.TxOut[i].Value
		fmt.Printf(" %15.8f BTC to %s\n", float64(tx.TxOut[i].Value)/1e8,
			btc.NewAddrFromPkScript(tx.TxOut[i].Pk_script, AddrVersion).String())
	}
	if missinginp {
		fmt.Println("WARNING: There are missing inputs and we cannot calc input BTC amount.")
		fmt.Println("If there is somethign wrong with this transaction, you can loose money...")
	} else {
		fmt.Printf("%.8f BTC in -> %.8f BTC out, with %.8f BTC fee\n", float64(totinp)/1e8,
			float64(totout)/1e8, float64(totinp-totout)/1e8)
	}
	TransactionsToSend[txid.Hash] = txd
	fmt.Println("Transaction stored in the memory pool. Please doeble check what it does.")
	fmt.Println("If it does what you needed, execute: stx " + txid.String())
}


func send_tx(par string) {
	txid := btc.NewUint256FromString(par)
	if txid==nil {
		fmt.Println("You must specify a valid transaction ID for this command.")
		list_txs("")
		return
	}
	if _, ok := TransactionsToSend[txid.Hash]; !ok {
		fmt.Println("No such transaction ID in the memory pool.")
		list_txs("")
		return
	}
	cnt := NetSendInv(1, txid.Hash[:], nil)
	fmt.Println("Transaction", txid.String(), "broadcasted to", cnt, "node(s)")
	fmt.Println("If it does not appear in the chain, you may want to redo it.")
}


func del_tx(par string) {
	txid := btc.NewUint256FromString(par)
	if txid==nil {
		fmt.Println("You must specify a valid transaction ID for this command.")
		list_txs("")
		return
	}
	if _, ok := TransactionsToSend[txid.Hash]; !ok {
		fmt.Println("No such transaction ID in the memory pool.")
		list_txs("")
		return
	}
	delete(TransactionsToSend, txid.Hash)
	fmt.Println("Transaction", txid.String(), "removed from the memory pool")
}


func list_txs(par string) {
	fmt.Println("Transactions in the memory pool:")
	cnt := 0
	for k, v := range TransactionsToSend {
		fmt.Println(cnt, btc.NewUint256(k[:]).String(), "-", len(v), "bytes")
	}
}

func init () {
	newUi("loadtx tx", true, load_tx, "Load transaction data from the given file, decode it and store in memory")
	newUi("sendtx stx", true, send_tx, "Broadcast transaction from memory pool (identified by a given <txid>)")
	newUi("deltx dtx", true, del_tx, "Temove a transaction from memory pool (identified by a given <txid>)")
	newUi("listtx ltx", true, list_txs, "List all the transaction loaded into memory pool")
}