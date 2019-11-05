import { Component } from '@angular/core';
import {DefaultWireGuardService} from './rpc/twirp_rpc';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss']
})
export class AppComponent {
  title = 'webapp';

  constructor() {
    const service = new DefaultWireGuardService('http://localhost:9090', window.fetch.bind(window));
    service.listPeers({offset: '1', limit: '2'})
      .then(response => {

        console.log('Got response', response);
      });
  }
}
