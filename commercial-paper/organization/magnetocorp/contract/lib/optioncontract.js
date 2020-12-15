/*
SPDX-License-Identifier: Apache-2.0
*/

'use strict';

// Fabric smart contract classes
const { Contract, Context } = require('fabric-contract-api');

// optionNet specifc classes
const ChargingOption = require('./option.js');
const OptionList = require('./optionlist.js');

/**
 * A custom context provides easy access to list of all charging options
 */
class ChargingOptionContext extends Context {

    constructor() {
        super();
        // All options are held in a list of options
        this.OptionList = new OptionList(this);
    }

}

/**
 * Define charging option smart contract by extending Fabric Contract class
 *
 */
class ChargingOptionContract extends Contract {

    constructor() {
        // Unique name when multiple contracts per chaincode file
        super('org.optionnet.chargingoption');
    }

    /**
     * Define a custom context for charging option
    */
    createContext() {
        return new ChargingOptionContext();
    }

    /**
     * Instantiate to perform any setup of the ledger that might be required.
     * @param {Context} ctx the transaction context
     */
    async instantiate(ctx) {
        // No implementation required with this example
        // It could be where data migration is performed, if necessary
        console.log('Instantiate the option contract');
    }

    /**
     * Issue charging option
     *
     * @param {Context} ctx the transaction context
     * @param {String} chargingStation charging option chargingStation
     * @param {Integer} optionNumber option number for this chargingStation
     * @param {String} issueDateTime option issue date
     * @param {String} maturityDateTime option maturity date
     * @param {Integer} faceValue face value of option
    */
    async issue(ctx, chargingStation, optionNumber, issueDateTime, maturityDateTime, faceValue) {

        // create an instance of the option
        let option = ChargingOption.createInstance(chargingStation, optionNumber, issueDateTime, maturityDateTime, faceValue);

        // Smart contract, rather than option, moves option into ISSUED state
        option.setIssued();

        // Newly issued option is owned by the chargingStation
        option.setOwner(chargingStation);

        // Add the option to the list of all similar charging options in the ledger world state
        await ctx.OptionList.addOption(option);

        // Must return a serialized option to caller of smart contract
        return option;
    }

    /**
     * Buy charging option
     *
     * @param {Context} ctx the transaction context
     * @param {String} chargingStation charging option chargingStation
     * @param {Integer} optionNumber option number for this chargingStation
     * @param {String} currentOwner current owner of option
     * @param {String} newOwner new owner of option
     * @param {Integer} price price paid for this option
     * @param {String} purchaseDateTime time option was purchased (i.e. traded)
    */
    async buy(ctx, chargingStation, optionNumber, currentOwner, newOwner, price, purchaseDateTime) {

        // Retrieve the current option using key fields provided
        let optionKey = ChargingOption.makeKey([chargingStation, optionNumber]);
        let option = await ctx.OptionList.getOption(optionKey);

        // Validate current owner
        if (option.getOwner() !== currentOwner) {
            throw new Error('option ' + chargingStation + optionNumber + ' is not owned by ' + currentOwner);
        }

        // First buy moves state from ISSUED to TRADING
        if (option.isIssued()) {
            option.setTrading();
        }

        // Check option is not already DELIVERED
        if (option.isTrading()) {
            option.setOwner(newOwner);
        } else {
            throw new Error('option ' + chargingStation + optionNumber + ' is not trading. Current state = ' +option.getCurrentState());
        }

        // Update the option
        await ctx.OptionList.updateOption(option);
        return option;
    }

    /**
     * Deliver charging option
     *
     * @param {Context} ctx the transaction context
     * @param {String} chargingStation charging option chargingStation
     * @param {Integer} optionNumber option number for this chargingStation
     * @param {String} deliveringOwner delivering owner of option
     * @param {String} deliverDateTime time option was delivered
    */
    async deliver(ctx, chargingStation, optionNumber, deliveringOwner, deliverDateTime) {

        let optionKey = ChargingOption.makeKey([chargingStation, optionNumber]);

        let option = await ctx.OptionList.getOption(optionKey);

        // Check option is not DELIVERED
        if (option.isDelivered()) {
            throw new Error('option ' + chargingStation + optionNumber + ' already delivered');
        }

        // Verify that the deliverer owns the charging option before delivering it
        if (option.getOwner() === deliveringOwner) {
            option.setOwner(option.getChargingStation());
            option.setDelivered();
        } else {
            throw new Error('Delivering owner does not own option' + chargingStation + optionNumber);
        }

        await ctx.OptionList.updateOption(option);
        return option;
    }

}

module.exports = ChargingOptionContract;
