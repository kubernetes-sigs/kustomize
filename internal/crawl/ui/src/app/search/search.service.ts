import { SearchResults } from '../documents';

import { Injectable } from '@angular/core';
import {
  HttpClient,
  HttpResponse,
  HttpParams } from '@angular/common/http';
import { Params, convertToParamMap } from '@angular/router';

import { Observable } from 'rxjs';
import { filter, map, catchError } from 'rxjs/operators';

@Injectable()
export class SearchService {
  private serviceUrl = "https://www.example.com/";

  constructor(private http: HttpClient) {}

  public search(params: Params): Observable<SearchResults> {
    let requestParams = new HttpParams();
    let pmap = convertToParamMap(params);
    let hasQuery = false;

    for (var k of pmap.keys) {
      for (var v of pmap.getAll(k)) {
        if (k == "q" && v != "") {
          hasQuery = true
        }
        requestParams.append(k, v)
      }
    }

    let queryUrl = this.serviceUrl
    if (hasQuery) {
      queryUrl += "search"
    } else {
      queryUrl += "metrics"
    }
    return this.http.get<SearchResults>(queryUrl, {params: params});
  }
}
