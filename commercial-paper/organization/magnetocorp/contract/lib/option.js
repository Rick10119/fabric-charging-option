/*
SPDX-License-Identifier: Apache-2.0
*/

'use strict';

// Utility class for ledger state
const State = require('../ledger-api/state.js');

// Enumerate charing option state values
const coState = {
    ISSUED: 1,
    TRADING: 2,
    DELIVERED: 3
};

/**
 * CharingOption class extends State class
 * Class will be used by application and smart contract to define a option
 */
class CharingOption extends State {

    constructor(obj) {
        super(CharingOption.getClass(), [obj.chargingStation, obj.optionNumber]);
        Object.assign(this, obj);
    }

    /**
     * Basic getters and setters
    */
    getChargingStation() {
        return this.chargingStation;
    }

    setChargingStation(newChargingStation) {
        this.chargingStation = newChargingStation;
    }

    getOwner() {
        return this.owner;
    }

    setOwner(newOwner) {
        this.owner = newOwner;
    }

    /**
     * Useful methods to encapsulate charging option states
     */
    setIssued() {
        this.currentState = coState.ISSUED;
    }

    setTrading() {
        this.currentState = coState.TRADING;
    }

    setDelivered() {
        this.currentState = coState.DELIVERED;
    }

    isIssued() {
        return this.currentState === coState.ISSUED;
    }

    isTrading() {
        return this.currentState === coState.TRADING;
    }

    isDelivered() {
        return this.currentState === coState.DELIVERED;
    }

    static fromBuffer(buffer) {
        return CharingOption.deserialize(buffer);
    }

    toBuffer() {
        return Buffer.from(JSON.stringify(this));
    }

    /**
     * Deserialize a state data to charging option
     * @param {Buffer} data to form back into the object
     */
    static deserialize(data) {
        return State.deserializeClass(data, CharingOption);
    }

    /**
     * Factory method to create a charging option object
     */
    static createInstance(chargingStation, optionNumber, issueDateTime, maturityDateTime, faceValue) {
        return new CharingOption({ chargingStation, optionNumber, issueDateTime, maturityDateTime, faceValue });
    }

    static getClass() {
        return 'org.optionnet.chargingoption';
    }
}

module.exports = CharingOption;
