import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { HistogramComponent } from './histogram.component';

describe('HistogramComponent', () => {
  let component: HistogramComponent;
  let fixture: ComponentFixture<HistogramComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ HistogramComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(HistogramComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
