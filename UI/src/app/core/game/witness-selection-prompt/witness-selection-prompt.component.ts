import { Component, OnDestroy, OnInit } from '@angular/core';
import { Subscription } from 'rxjs';
import { GameApiService } from '../../../shared/api/game/game-api.service';
import { TgKnownRolePlayer, TgPlayer, TgPlayerPrivateData, TgWitnessSelectionPromptState } from '../../../shared/api/models/models';
import { SnackBarService } from '../../../shared/api/snack-bar/snack-bar.service';

@Component({
  selector: 'app-witness-selection-prompt',
  standalone: false,
  templateUrl: './witness-selection-prompt.component.html',
  styleUrls: ['./witness-selection-prompt.component.scss'],
})
export class WitnessSelectionPromptComponent implements OnInit, OnDestroy {
  prompt: TgWitnessSelectionPromptState;
  players: TgPlayer[] = [];
  privateData: TgPlayerPrivateData;
  minimized = false;
  selectedUids: string[] = [];
  private subscription = new Subscription();

  constructor(
    public gameApi: GameApiService,
    private snack: SnackBarService,
  ) {}

  ngOnInit() {
    this.subscription.add(
      this.gameApi.playerPrivateData$.subscribe((privateData) => {
        this.privateData = privateData;
        this.prompt = privateData ? privateData.activeWitnessSelectionPrompt : null;
        if (!this.prompt) {
          this.selectedUids = [];
          this.minimized = false;
        }
      }),
    );
    this.subscription.add(
      this.gameApi.players$.subscribe((players) => {
        this.players = players || [];
      }),
    );
  }

  ngOnDestroy(): void {
    this.subscription.unsubscribe();
  }

  get visible() {
    return !!(this.prompt && this.prompt.active);
  }

  get candidates() {
    const murdererTeam = new Set((this.privateData?.knownMurdererTeam || []).map((player: TgKnownRolePlayer) => player.uid));
    return this.players.filter((player) => !murdererTeam.has(player.uid));
  }

  toggleSelected(uid: string) {
    if (this.selectedUids.includes(uid)) {
      this.selectedUids = this.selectedUids.filter((selectedUid) => selectedUid !== uid);
      return;
    }
    if (this.selectedUids.length >= this.prompt.requiredSelections) {
      this.snack.error(`Select exactly ${this.prompt.requiredSelections} player(s).`);
      return;
    }
    this.selectedUids = [...this.selectedUids, uid];
  }

  async submit() {
    if (this.selectedUids.length !== this.prompt.requiredSelections) {
      this.snack.error(`Select exactly ${this.prompt.requiredSelections} player(s).`);
      return;
    }
    await this.gameApi.submitWitnessSelection(this.selectedUids);
    this.selectedUids = [];
  }

  async dismiss() {
    await this.gameApi.dismissWitnessSelectionPrompt();
    this.selectedUids = [];
  }

  isSelected(uid: string) {
    return this.selectedUids.includes(uid);
  }
}
