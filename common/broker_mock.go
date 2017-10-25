/*
 * Copyright Morpheo Org. 2017
 *
 * contact@morpheo.co
 *
 * This software is part of the Morpheo project, an open-source machine
 * learning platform.
 *
 * This software is governed by the CeCILL license, compatible with the
 * GNU GPL, under French law and abiding by the rules of distribution of
 * free software. You can  use, modify and/ or redistribute the software
 * under the terms of the CeCILL license as circulated by CEA, CNRS and
 * INRIA at the following URL "http://www.cecill.info".
 *
 * As a counterpart to the access to the source code and  rights to copy,
 * modify and redistribute granted by the license, users are provided only
 * with a limited warranty  and the software's author,  the holder of the
 * economic rights,  and the successive licensors  have only  limited
 * liability.
 *
 * In this respect, the user's attention is drawn to the risks associated
 * with loading,  using,  modifying and/or developing or reproducing the
 * software by the user in light of its specific status of free software,
 * that may mean  that it is complicated to manipulate,  and  that  also
 * therefore means  that it is reserved for developers  and  experienced
 * professionals having in-depth computer knowledge. Users are therefore
 * encouraged to load and test the software's suitability as regards their
 * requirements in conditions enabling the security of their systems and/or
 * data to be ensured and,  more generally, to use and operate it in the
 * same conditions as regards security.
 *
 * The fact that you are presently reading this means that you have had
 * knowledge of the CeCILL license and that you accept its terms.
 */

package common

import (
	"time"
)

const (
	// BrokerMOCK identifies the MOCK broker type among other brokers (used when the user specifies the
	// broker to be used as a CLI flag)
	BrokerMOCK = "mock"
)

// ProducerMOCK is an implementation of our Producer interface for MOCK
type ProducerMOCK struct {
}

// Push returns nil
func (p *ProducerMOCK) Push(topic string, body []byte) (err error) {
	return nil
}

// Stop returns nothing
func (p *ProducerMOCK) Stop() {
	return
}

// ConsumerMOCK implements an MOCK version of our Consumer interface
type ConsumerMOCK struct {
}

// ConsumeUntilKilled listens for messages on a given MOCK (topic, channel) pair until it's killed
func (c *ConsumerMOCK) ConsumeUntilKilled() {
	return
}

// AddHandler adds a handler function (with a tunable level of concurrency) to our MOCK consumer
func (c *ConsumerMOCK) AddHandler(topic string, handler Handler, concurrency int, timeout time.Duration) (err error) {
	return nil
}
