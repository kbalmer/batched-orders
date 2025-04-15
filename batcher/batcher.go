package batcher

import (
	"log"
	"strconv"
	"sync"
	"time"

	"batched-orders/pkg"
	"batched-orders/sftp"
)

type Batcher struct {
	batchSize      int
	batchCh        chan pkg.Order
	store          map[string]pkg.Order // TODO: replace with redis to make this service stateless
	mu             sync.RWMutex
	sendCh         chan pkg.Order
	SFTP           sftp.SFTP
	downloadTicker *time.Ticker
	doneCh         chan struct{}
}

func NewBatcher(batchSize int, downloadTicker *time.Ticker, sftp sftp.SFTP) (*Batcher, chan struct{}) {
	doneCh := make(chan struct{})
	return &Batcher{
		batchSize:      batchSize,
		batchCh:        make(chan pkg.Order, batchSize),
		store:          make(map[string]pkg.Order, batchSize),
		mu:             sync.RWMutex{},
		sendCh:         make(chan pkg.Order, batchSize),
		SFTP:           sftp,
		downloadTicker: downloadTicker,
		doneCh:         doneCh,
	}, doneCh
}

func (b *Batcher) Batch(order pkg.Order) {
	b.batchCh <- order
}

func (b *Batcher) BatchListener() {
	for {
		select {
		case order := <-b.batchCh:
			if len(b.store) < b.batchSize {
				b.put(order)
			} else {
				b.drain()
			}
		case <-b.doneCh:
			return
		default:
		}
	}
}

func generateKey(order pkg.Order) string {
	return strconv.FormatInt(order.Timestamp.UnixMilli(), 0) + "_" + strconv.FormatInt(order.Customer.CustomerNumber, 0)
}

func (b *Batcher) put(order pkg.Order) {
	defer b.mu.Unlock()
	b.mu.Lock()

	b.store[generateKey(order)] = order
}

func (b *Batcher) drain() {
	for key, order := range b.store {
		b.sendCh <- order
		delete(b.store, key)
	}
}

func (b *Batcher) updateOrderState(key string, newState pkg.OrderState) {
	defer b.mu.Unlock()
	b.mu.Lock()

	// get current order and check state
	currentOrder := b.store[key]
	// if state is different update the state to the new state
	if currentOrder.State != newState {
		currentOrder.State = newState
		b.store[key] = currentOrder
	}
}

func (b *Batcher) Send() {
	for {
		select {
		case <-b.doneCh:
			return
		default:
		}

		if len(b.sendCh) == b.batchSize {
			var ordersToUpload = make([]pkg.Order, b.batchSize)

			for order := range b.sendCh {
				ordersToUpload = append(ordersToUpload, order)
			}

			if err := b.SFTP.Upload(ordersToUpload); err != nil {
				log.Printf("Error uploading orders to SFTP: %v", err)
			}

			for _, orderToBeUpdate := range ordersToUpload {
				b.updateOrderState(generateKey(orderToBeUpdate), pkg.OrderStateSent)
			}
		}
	}
}

func (b *Batcher) GetLatest() {
	for {
		select {
		case <-b.downloadTicker.C:
			var ordersToBeDownloaded []pkg.Order
			for _, order := range b.store {
				if order.State == pkg.OrderStateSent {
					ordersToBeDownloaded = append(ordersToBeDownloaded, order)
				}
			}
			if err := b.SFTP.Download(ordersToBeDownloaded); err != nil {
				log.Printf("Error downloading orders to SFTP: %v", err)
			}

		case <-b.doneCh:
			return
		}
	}
}
