import { Chart } from 'chart.js';
import { SearchResults } from '../documents';

import { Component, OnInit } from '@angular/core';
import { Subject, Observable } from 'rxjs';

@Component({
  selector: 'app-timeseries',
  templateUrl: './timeseries.component.html',
  styleUrls: ['./timeseries.component.css']
})
export class TimeseriesComponent implements OnInit {
  timeseries;

  constructor() {}

  ngOnInit() {}

  update(agg: SearchResults.BucketAggregation) {
    if (this.timeseries) {
      this.timeseries.destroy();
    }
    if (!agg || agg.buckets.length == 0) {
      this.timeseries = null;
      return
    }

    let buckets = agg.buckets
      .filter(bucket => new Date(bucket.key) > new Date(2017, 1));

    let labels = buckets.map(bucket => new Date(bucket.key))
    let counts = buckets.map(bucket => bucket.count);

    let sum = 0;
    for (let i = 0; i < counts.length; i++) {
      sum += counts[i];
      counts[i] = sum;
    }

    this.timeseries = new Chart('timeseries', {
      type: 'line',
      data: {
        datasets: [{
          label: 'Kustomizations Over time',
          data: counts,
          type: 'line',
          pointRadius: 0,
          lineTension: 0,
        }],
        labels: labels,
      },
      options: {
        scales: {
          xAxes: [{
            type: 'time',
            distribution: 'linear',
            ticks: {
              autoSkip: true,
            },
          }],
        }
      },
    })
  }
}
