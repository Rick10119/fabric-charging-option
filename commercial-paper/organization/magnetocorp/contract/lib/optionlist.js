/*
SPDX-License-Identifier: Apache-2.0
*/

'use strict';

// Utility class for collections of ledger states --  a state list
const StateList = require('../ledger-api/statelist.js');

const ChargingOption = require('./option.js');

class OptionList extends StateList {

    constructor(ctx) {
        super(ctx, 'org.optionnet.chargingoptionlist');
        this.use(ChargingOption);
    }

    async addOption(option) {
        return this.addState(option);
    }

    async getOption(optionKey) {
        return this.getState(optionKey);
    }

    async updateOption(option) {
        return this.updateState(option);
    }
}


module.exports = OptionList;