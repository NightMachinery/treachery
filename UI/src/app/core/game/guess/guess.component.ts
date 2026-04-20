import { Component, Input, OnInit } from '@angular/core';
import { take } from 'rxjs/operators';
import { ChatApiService } from 'src/app/shared/api/chat/chat-api.service';
import { GameApiService } from './../../../shared/api/game/game-api.service';
import { TgPartialGuess, TgPlayer } from '../../../shared/api/models/models';
import { SnackBarService } from 'src/app/shared/api/snack-bar/snack-bar.service';

@Component({
  selector: 'app-guess',
  standalone: false,
  templateUrl: './guess.component.html',
  styleUrls: ['./guess.component.scss']
})
export class GuessComponent implements OnInit {
  @Input() guess: TgPartialGuess;

  constructor(public gameApi: GameApiService, public snack: SnackBarService, public chatApi: ChatApiService) {}

  ngOnInit() {}

  makeGuess() {
    this.gameApi.guesses$.pipe(take(1)).subscribe(guesses => {
      this.gameApi.playersDict$.pipe(take(1)).subscribe(playersDict => {
        this.gameApi.viewer$.pipe(take(1)).subscribe(viewer => {
          const duplicateGuess = guesses.some(
            guess =>
              guess.murdererUid === this.guess.murdererUid &&
              guess.meansCardName === this.guess.meansCardName &&
              guess.clueCardName === this.guess.clueCardName
          );
          if (duplicateGuess) {
            this.snack.error('That exact guess was already submitted.');
            return;
          }
          if (guesses.filter(guess => guess.guessedByUid === viewer.uid).length === 0) {
            this.gameApi.makeGuess(this.guess);
          } else {
            this.snack.error("You've already made a guess! But will send a message for your investigator colleagues.");
            this.chatApi.sendMessage(
              `I don't have any guesses left, but I think it's ${playersDict.get(this.guess.murdererUid).name} with '${
                this.guess.clueCardName
              }' and '${this.guess.meansCardName}'`
            );
          }
        });
      });
    });
  }

  getMurderer(players) {
    return players.find(player => player.uid === this.guess.murdererUid);
  }

  getClueCard(murderer: TgPlayer) {
    return murderer.clueCards.find(card => card.name === this.guess.clueCardName);
  }

  getMeansCard(murderer: TgPlayer) {
    return murderer.meansCards.find(card => card.name === this.guess.meansCardName);
  }
}
