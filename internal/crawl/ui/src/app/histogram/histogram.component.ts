import { Chart } from 'chart.js';
import { SearchResults } from '../documents';

import { Component, OnInit } from '@angular/core';
import { Subject, Observable } from 'rxjs';

const otherLabel = 'Other Kinds';

// Draws a histogram from SearchResults.BucketAggregation data.
@Component({
  selector: 'app-histogram',
  templateUrl: './histogram.component.html',
  styleUrls: ['./histogram.component.css']
})
export class HistogramComponent implements OnInit {
  hist;

  constructor() {}
  ngOnInit() {}

  public update(agg: SearchResults.BucketAggregation): Observable<string> {
    if (this.hist) {
      this.hist.destroy();
    }

    let labels = agg.buckets.map(bucket => bucket.key);
    let counts = agg.buckets.map(bucket => bucket.count);
    if (agg.otherResults && agg.otherResults > 0) {
      labels.push(otherLabel)
      counts.push(agg.otherResults)
    }

    let selectedLabel = new Subject<string>();

    this.hist = new Chart('histogram', {
      type: 'bar',
      data: {
        datasets: [ { data: counts } ],
        labels: labels,
      },
      options: {
        legend: { display: false },
        'onClick' : function(e, it) {
          if (!(it && it[0] && it[0]._model && it[0]._model.label)) {
            return
          }
          let label = it[0]._model.label;
          if (label != otherLabel) {
            selectedLabel.next(label);
          }
        }.bind(selectedLabel),
        scales: {
          // no floating point
          yAxes: [ { ticks: { precision: 0, beginAtZero: true } } ],
        },
      },
    })

    return selectedLabel;
  }
}
