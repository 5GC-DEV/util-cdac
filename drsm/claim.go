// SPDX-FileCopyrightText: 2022 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0
package drsm

import (
	"fmt"

	"github.com/5GC-DEV/util-cdac/logger"
	"go.mongodb.org/mongo-driver/bson"
)

func (d *Drsm) podDownDetected() {
	logger.DrsmLog.Infoln("started Pod Down goroutine")

	for p := range d.podDown {
		logger.DrsmLog.Infof("pod Down detected %v", p)

		// Safely get pod entry
		d.podMapMutex.RLock()
		pd, found := d.podMap[p]
		d.podMapMutex.RUnlock()

		if !found || pd == nil {
			logger.DrsmLog.Warnf("pod %s not found in podMap", p)
			continue
		}

		// Copy chunk IDs while holding read lock
		pd.mu.RLock()
		chunkIDs := make([]int32, 0, len(pd.podChunks))
		for k := range pd.podChunks {
			chunkIDs = append(chunkIDs, k)
		}
		pd.mu.RUnlock()

		// Process chunks without holding pod lock
		for _, k := range chunkIDs {

			d.globalChunkTblMutex.RLock()
			c, found := d.globalChunkTbl[k]
			d.globalChunkTblMutex.RUnlock()

			logger.DrsmLog.Debugf("found: %v chunk: %v", found, c)

			if found && c != nil {
				go c.claimChunk(d, pd.PodId.PodName)
			}
		}
	}
}

func (c *chunk) claimChunk(d *Drsm, curOwner string) {
	// Need optimization
	if d.mode != ResourceClient {
		logger.DrsmLog.Infoln("claimChunk ignored demux mode")
		return
	}
	// try to claim. If success then notification will update owner.
	logger.DrsmLog.Debugln("claimChunk started")
	docId := fmt.Sprintf("chunkid-%d", c.Id)
	update := bson.M{"_id": docId, "type": "chunk", "podId": d.clientId.PodName, "podInstance": d.clientId.PodInstance, "podIp": d.clientId.PodIp}
	filter := bson.M{"_id": docId, "podId": curOwner}
	updated := d.mongo.RestfulAPIPutOnly(d.sharedPoolName, filter, update)
	if updated == nil {
		// TODO : don't add to local pool yet. We can add it only if scan is done.
		logger.DrsmLog.Infof("claimChunk %v success", c.Id)
		c.Owner.PodName = d.clientId.PodName
		c.Owner.PodIp = d.clientId.PodIp
		go c.scanChunk(d)
	} else {
		// no problem, some other POD successfully claimed this chunk
		logger.DrsmLog.Infof("claimChunk %v failure", c.Id)
	}
}
