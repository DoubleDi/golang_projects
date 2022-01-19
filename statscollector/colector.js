/*
* Our application servers receive approximately 20 000
* http requests  per second. Response timeout is 19000ms.
* Implement a statistics collector that calculates the
* median and average request response times for a 7 day
* dataset.
*
* Assigment:
* 1. Implement StatsCollector
* 2. Write tests (below StatsCollector)
*/

'use strict';

// 7 * 24 * 3600 * 20000 * 19000 = 229824000000000 fits in number type
/* Mutex? */
// State collector
class StatsCollector {
	constructor(/*void*/) {
  	this.sum = 0;
    this.count = 0;
    this.timings = Array.from({length: 19000}, (_, i) => 0);
  }

	pushValue(responseTimeMs /*number*/) /*void*/ {
    this.sum += responseTimeMs;
    this.count++;
    this.timings[responseTimeMs]++;
  }

	getMedian() /*number*/ {
  	if (this.count == 0) {
    	return 0;
    }
    var half = Math.floor(this.count / 2);
    var zeroIndex = undefined;
    for (var i = 0; i < this.timings.length; i++) {
       half -= this.timings[i];
       if (half < 0) {
       	 zeroIndex = zeroIndex ? zeroIndex : i
         return this.count % 2 == 0 ? (i + zeroIndex) / 2 : i;
       }
       if (half == 0 && zeroIndex == undefined) {
       	 zeroIndex = i;
       }
    }
  }

	getAverage() /*number*/ {
  	return this.count > 0 ? this.sum / this.count : 0;
  }

}

// Configure Mocha, telling both it and chai to use BDD-style tests.
mocha.setup("bdd");
chai.should();

describe('StatsCollector', function(){
  it('[1,1,3,6,6]', function(){
  	var c = new StatsCollector();
    c.pushValue(0);
    c.pushValue(0);
    c.pushValue(3);
    c.pushValue(19000);
    c.pushValue(19000);
    c.getAverage().should.equal(7600.6);
    c.getMedian().should.equal(3);
  });
  it('[]', function(){
  	var c = new StatsCollector();
    c.getAverage().should.equal(0);
    c.getMedian().should.equal(0);
  });
  it('[1,1,3,3]', function(){
  	var c = new StatsCollector();
    c.pushValue(1);
    c.pushValue(1);
    c.pushValue(3);
    c.pushValue(3);
    c.getAverage().should.equal(2);
    c.getMedian().should.equal(2);
  });
  it('[1,1,1,1,3,2]', function(){
  	var c = new StatsCollector();
    c.pushValue(1);
    c.pushValue(1);
    c.pushValue(1);
    c.pushValue(1);
    c.pushValue(3);
    c.pushValue(2);
    c.getAverage().should.equal(1.5);
    c.getMedian().should.equal(1);
  });
  it('[6,6,1,1]', function(){
  	var c = new StatsCollector();
    c.pushValue(6);
    c.pushValue(6);
    c.pushValue(1);
    c.pushValue(1);
    c.getAverage().should.equal(3.5);
    c.getMedian().should.equal(3.5);
  });
});

// Run all our test suites.  Only necessary in the browser.
mocha.run();
    