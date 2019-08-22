import { BrowserModule } from '@angular/platform-browser';
import { Routes, RouterModule } from '@angular/router';
import { NgModule } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { HttpClientModule } from '@angular/common/http';

import { MatExpansionModule } from '@angular/material/expansion';
import { MatInputModule } from '@angular/material/input';
import { MatListModule } from '@angular/material/list';
import { MatButtonModule } from '@angular/material/button';

import { AppComponent } from './app.component';
import { SearchComponent } from './search/search.component';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { HistogramComponent } from './histogram/histogram.component';
import { TimeseriesComponent } from './timeseries/timeseries.component';

const appRoutes: Routes = [
  {
    path: 'search',
    component: SearchComponent,
    runGuardsAndResolvers: 'always'
  },
  // Always ridirect to the search endpoint for now.
  {
    path: '',
    redirectTo: 'search',
    pathMatch: 'full',
  },
];

@NgModule({
  declarations: [
    AppComponent,
    SearchComponent,
    HistogramComponent,
    TimeseriesComponent,
  ],
  imports: [
    BrowserModule,
    BrowserAnimationsModule,
    HttpClientModule,
    MatExpansionModule,
    MatInputModule,
    MatListModule,
    MatButtonModule,
    FormsModule,
    RouterModule.forRoot(
      appRoutes,
      { onSameUrlNavigation: 'reload', }
    )
  ],
  providers: [
    {provide: HttpClientModule}
  ],
  bootstrap: [AppComponent]
})
export class AppModule {}
