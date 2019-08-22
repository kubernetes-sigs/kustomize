import { Component, OnInit, NgModule } from '@angular/core';
import { Router, ActivatedRoute } from '@angular/router';
import { SearchResults } from '../documents';
import { HistogramComponent } from '../histogram/histogram.component';
import { TimeseriesComponent } from '../timeseries/timeseries.component';
import { SearchService } from './search.service';

const perPage = 10;

@Component({
  selector: 'app-search',
  templateUrl: './search.component.html',
  styleUrls: ['./search.component.css'],
  providers: [SearchService]
})
export class SearchComponent implements OnInit {
  inputQuery: string[] = [];
  from: number = 0;
  disableNav: boolean = false;

  docs: SearchResults = {
    hits: {
      total: 0,
      hits: [],
    },
  };

  kindBreakdown = new HistogramComponent();
  timeseries = new TimeseriesComponent();

  constructor(
    private searcher : SearchService,
    private router: Router,
    private route: ActivatedRoute
  ) {}

  ngOnInit() {
    this.route.queryParams.subscribe(params => {
      if (params.q instanceof Array) {
        this.inputQuery = params.q || [""]
      } else {
        this.inputQuery = [params.q || "" ];
      }

      this.from = parseInt(params.from) || 0;
      if (this.from < 0) {
        this.from = Math.max(this.from, 0);
        this.searchWithParams();
      }

      this.searcher.search(params).subscribe(sr => {
        this.docs = sr;
        this.kindBreakdown.update(sr.aggregations.kinds).subscribe(selectedKind => {
          this.addToQuery('kind='+selectedKind)
          this.search();
        })
        this.timeseries.update(sr.aggregations.timeseries);
      });
    });
  }

  public addToQuery(q: string) {
    for (let v of this.inputQuery) {
      if (v == q) {
        return
      }
    }
    this.inputQuery.push(q)
  }

  search(): void {
    this.from = 0;
    this.searchWithParams();
  }

  searchWithParams(): void {
    let params = {
      q: this.inputQuery,
      from: this.from,
    }
    this.router.navigate([], {
      relativeTo: this.route,
      queryParams: params,
    });
  }

  first(): boolean {
    return this.from <= 0 || this.disableNav;
  }

  last(): boolean {
    return this.from + perPage >= this.docs.hits.total || this.disableNav;
  }

  next (): void {
    this.from += perPage;
    this.searchWithParams();
  }
  prev (): void {
    this.from -= perPage;
    this.searchWithParams();
  }

  get inputQueryValue() : string {
    return this.inputQuery.join(' ')
  }

  set inputQueryValue(input : string) {
    this.inputQuery = [input]
  }
}
