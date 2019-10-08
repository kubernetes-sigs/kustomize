export interface SearchResults {
  hits: SearchResults.Hits;
  aggregations?: SearchResults.Aggregations;
};

export namespace SearchResults {
  export class Hits {
    total: number;
    hits: SearchResults.InnerHits[];
  };

  export class InnerHits {
    id: string;
    result: SearchResults.Result;
  };

  export class Result {
    repositoryUrl: string;
    filePath: string;
    defaultBranch: string;
    document: string;
    creationTime: Date;
    values: string;
    kinds: string;
  };

  export interface Aggregations {
    timeseries?: SearchResults.BucketAggregation;
    kinds?: SearchResults.BucketAggregation;
  };

  export interface BucketAggregation {
    otherResults?: number;
    buckets: SearchResults.Bucket[];
  };

  export class Bucket {
    key: string;
    count: number;
  };
};
