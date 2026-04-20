import { Component, OnInit, Input } from '@angular/core';

@Component({
  selector: 'tg-info',
  standalone: false,
  templateUrl: './info.component.html',
  styleUrls: ['./info.component.scss']
})
export class InfoComponent implements OnInit {
  @Input() text: string;
  constructor() { }

  ngOnInit(): void {
  }

}
