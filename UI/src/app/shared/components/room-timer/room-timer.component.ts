import { Component, OnDestroy, OnInit } from '@angular/core';
import { Subscription } from 'rxjs';
import { GameApiService } from '../../api/game/game-api.service';
import { TgGameSnapshot, TgRoomTimer } from '../../api/models/models';
import { SnackBarService } from '../../api/snack-bar/snack-bar.service';

const DEFAULT_ROOM_TIMER_SECONDS = 40;

type RoomTimerStatus = 'idle' | 'running' | 'paused' | 'expired';

@Component({
  selector: 'app-room-timer',
  standalone: false,
  templateUrl: './room-timer.component.html',
  styleUrls: ['./room-timer.component.scss'],
})
export class RoomTimerComponent implements OnInit, OnDestroy {
  durationSeconds = DEFAULT_ROOM_TIMER_SECONDS;
  remainingSeconds = 0;
  status: RoomTimerStatus = 'idle';
  roomTimer: TgRoomTimer = null;
  isCreator = false;
  busy = false;

  private subscription = new Subscription();
  private ticker: any = null;
  private serverOffsetMs = 0;

  constructor(
    public gameApi: GameApiService,
    private snack: SnackBarService,
  ) {}

  ngOnInit(): void {
    this.subscription.add(
      this.gameApi.snapshot$.subscribe((snapshot) => {
        this.syncFromSnapshot(snapshot);
      }),
    );
    this.startTicker();
  }

  ngOnDestroy(): void {
    this.subscription.unsubscribe();
    if (this.ticker) {
      clearInterval(this.ticker);
      this.ticker = null;
    }
  }

  get shouldRender(): boolean {
    return !!this.roomTimer || this.isCreator;
  }

  get isRunning(): boolean {
    return this.status === 'running';
  }

  get isPaused(): boolean {
    return this.status === 'paused';
  }

  get isExpired(): boolean {
    return this.status === 'expired';
  }

  get timeLabel(): string {
    const totalSeconds = Math.max(
      0,
      this.status === 'paused' && this.roomTimer ? this.roomTimer.pausedRemainingSeconds || 0 : this.remainingSeconds,
    );
    const minutes = Math.floor(totalSeconds / 60);
    const seconds = totalSeconds % 60;
    return `${minutes}:${seconds.toString().padStart(2, '0')}`;
  }

  normalizeDurationInput() {
    const parsed = Math.floor(Number(this.durationSeconds));
    this.durationSeconds = parsed >= 1 ? parsed : DEFAULT_ROOM_TIMER_SECONDS;
  }

  adjustDuration(delta: number) {
    this.durationSeconds = (Number(this.durationSeconds) || DEFAULT_ROOM_TIMER_SECONDS) + delta;
    this.normalizeDurationInput();
  }

  async startTimer() {
    this.normalizeDurationInput();
    await this.performAction(async () => this.gameApi.startRoomTimer(this.durationSeconds));
  }

  async pauseTimer() {
    await this.performAction(async () => this.gameApi.pauseRoomTimer());
  }

  async resumeTimer() {
    await this.performAction(async () => this.gameApi.resumeRoomTimer());
  }

  async resetTimer() {
    await this.performAction(async () => this.gameApi.resetRoomTimer());
  }

  async clearTimer() {
    await this.performAction(async () => this.gameApi.clearRoomTimer());
  }

  private syncFromSnapshot(snapshot: TgGameSnapshot) {
    this.roomTimer = snapshot && snapshot.game ? snapshot.game.roomTimer || null : null;
    this.isCreator = !!(snapshot && snapshot.viewer && snapshot.viewer.isCreator);
    const parsedServerTimestamp = snapshot && snapshot.serverTimestamp ? Date.parse(snapshot.serverTimestamp) : NaN;
    this.serverOffsetMs = Number.isNaN(parsedServerTimestamp) ? 0 : parsedServerTimestamp - Date.now();
    if (this.roomTimer) {
      this.durationSeconds = this.roomTimer.durationSeconds || this.durationSeconds || DEFAULT_ROOM_TIMER_SECONDS;
    }
    this.updateDerivedState();
  }

  private startTicker() {
    this.ticker = setInterval(() => this.updateDerivedState(), 250);
  }

  private updateDerivedState() {
    if (!this.roomTimer) {
      this.remainingSeconds = 0;
      this.status = 'idle';
      return;
    }

    if (this.roomTimer.expiresAt) {
      this.remainingSeconds = this.calculateRemainingSeconds(this.roomTimer.expiresAt);
      this.status = this.remainingSeconds > 0 ? 'running' : 'expired';
      if (this.status === 'expired') {
        this.maybeTriggerExpiryCue();
      }
      return;
    }

    if (this.roomTimer.pausedRemainingSeconds > 0) {
      this.remainingSeconds = this.roomTimer.pausedRemainingSeconds;
      this.status = 'paused';
      return;
    }

    this.remainingSeconds = 0;
    this.status = 'idle';
  }

  private calculateRemainingSeconds(expiresAt: string): number {
    const expiresAtMs = Date.parse(expiresAt);
    if (Number.isNaN(expiresAtMs)) {
      return 0;
    }
    const diffMs = expiresAtMs - (Date.now() + this.serverOffsetMs);
    if (diffMs <= 0) {
      return 0;
    }
    return Math.ceil(diffMs / 1000);
  }

  private maybeTriggerExpiryCue() {
    if (!this.roomTimer || !this.roomTimer.runId) {
      return;
    }
    const storageKey = this.getCueStorageKey(this.roomTimer.runId);
    let alreadyTriggered = false;
    try {
      alreadyTriggered = sessionStorage.getItem(storageKey) === '1';
      if (!alreadyTriggered) {
        sessionStorage.setItem(storageKey, '1');
      }
    } catch (error) {
      alreadyTriggered = false;
    }
    if (alreadyTriggered) {
      return;
    }
    this.playExpiryBeep();
  }

  private getCueStorageKey(runId: number): string {
    const gameId = this.gameApi.gameId$.value || 'unknown';
    return `treachery-room-timer-cue:${gameId}:${runId}`;
  }

  private async playExpiryBeep() {
    try {
      const AudioContextCtor = (window as any).AudioContext || (window as any).webkitAudioContext;
      if (!AudioContextCtor) {
        return;
      }
      const audioContext = new AudioContextCtor();
      if (audioContext.state === 'suspended' && audioContext.resume) {
        await audioContext.resume();
      }
      const oscillator = audioContext.createOscillator();
      const gainNode = audioContext.createGain();
      oscillator.type = 'sine';
      oscillator.frequency.value = 880;
      gainNode.gain.setValueAtTime(0.0001, audioContext.currentTime);
      gainNode.gain.exponentialRampToValueAtTime(0.2, audioContext.currentTime + 0.01);
      gainNode.gain.exponentialRampToValueAtTime(0.0001, audioContext.currentTime + 0.3);
      oscillator.connect(gainNode);
      gainNode.connect(audioContext.destination);
      oscillator.start();
      oscillator.stop(audioContext.currentTime + 0.3);
      oscillator.onended = () => {
        if (audioContext.close) {
          audioContext.close().catch(() => undefined);
        }
      };
    } catch (error) {
      // Visual cue remains even if sound playback is unavailable.
    }
  }

  private async performAction(action: () => Promise<void>) {
    if (this.busy) {
      return;
    }
    this.busy = true;
    try {
      await action();
    } catch (error) {
      const message = error && error.error && error.error.error ? error.error.error : 'Could not update the room timer.';
      this.snack.error(message);
    } finally {
      this.busy = false;
    }
  }
}
