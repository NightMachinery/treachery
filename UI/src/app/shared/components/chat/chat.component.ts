import { Component, ElementRef, Input, OnInit, ViewChild } from '@angular/core';
import { ChatApiService } from '../../api/chat/chat-api.service';
import { TgMessage, TgGame } from '../../api/models/models';
import { GameApiService } from './../../api/game/game-api.service';

@Component({
  selector: 'app-chat',
  templateUrl: './chat.component.html',
  styleUrls: ['./chat.component.scss']
})
export class ChatComponent implements OnInit {
  @Input() disableChat = false;
  @ViewChild('messages') messagesEl: ElementRef;
  message: string;

  constructor(public chatApi: ChatApiService, public gameApi: GameApiService) {}

  mine(message: TgMessage, viewerUid: string) {
    return !!viewerUid && message.playerUid === viewerUid;
  }

  ngOnInit() {}

  sendMessage() {
    if (!this.message || !this.message.trim()) {
      return;
    }
    this.chatApi.sendMessage(this.message.trim());
    this.message = '';
    this.scrollToBottom();
  }

  scrollToBottom() {
    setTimeout(() => {
      if (this.messagesEl && this.messagesEl.nativeElement) {
        this.messagesEl.nativeElement.scrollTop = this.messagesEl.nativeElement.scrollHeight;
      }
    }, 200);
  }
}
