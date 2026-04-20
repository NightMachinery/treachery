import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { HttpClientModule, HTTP_INTERCEPTORS } from '@angular/common/http';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { FormsModule } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatDividerModule } from '@angular/material/divider';
import { MatInputModule } from '@angular/material/input';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatRippleModule } from '@angular/material/core';
import { MatListModule } from '@angular/material/list';
import { MatDialogModule, MAT_DIALOG_DEFAULT_OPTIONS } from '@angular/material/dialog';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatSnackBarModule } from '@angular/material/snack-bar';
import { LazyLoadImageModule } from 'ng-lazyload-image';

import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { GameComponent } from './core/game/game.component';
import { ForensicComponent } from './core/forensic/forensic.component';
import { AllGamesComponent } from './core/all-games/all-games.component';
import { CardComponent } from './shared/components/card/card.component';
import { ActiveGamesListItemComponent } from './core/all-games/active-games-list-item/active-games-list-item.component';
import { JoinGameComponent } from './core/join-game/join-game.component';
import { JoinedPlayersListComponent } from './core/join-game/joined-players-list/joined-players-list.component';
import { JoinedPlayerListItemComponent } from './core/join-game/joined-player-list-item/joined-player-list-item.component';
import { RippleMatCardComponent } from './shared/components/ripple-mat-card/ripple-mat-card.component';
import { PlayerDeckComponent } from './core/game/player-deck/player-deck.component';
import { ChatComponent } from './shared/components/chat/chat.component';
import { PlayerDeckPagerComponent } from './core/game/player-deck-pager/player-deck-pager.component';
import { GuessComponent } from './core/game/guess/guess.component';
import { ForensicCardComponent } from './core/forensic/forensic-card/forensic-card.component';
import { MurdererSelectDialogComponent } from './core/game/murderer-select-dialog/murderer-select-dialog.component';
import { TestComponent } from './test/test.component';
import { CenteringFlexComponent } from './shared/components/centering-flex/centering-flex.component';
import { AvatarComponent } from './shared/components/avatar/avatar.component';
import { NavbarComponent } from './shared/components/navbar/navbar.component';
import { CopyLinkComponent } from './shared/copy-link/copy-link.component';
import { InfoComponent } from './shared/components/info/info.component';
import { MurdererInfoComponent } from './core/forensic/murderer-info/murderer-info.component';
import { ConfirmationButtonComponent } from './shared/components/confirmation-button/confirmation-button.component';
import { PromptPanelComponent } from './shared/components/prompt-panel/prompt-panel.component';
import { WitnessSelectionPromptComponent } from './core/game/witness-selection-prompt/witness-selection-prompt.component';
import { ModeratorWitnessControlsComponent } from './core/game/moderator-witness-controls/moderator-witness-controls.component';
import { PrivateRolePanelComponent } from './core/game/private-role-panel/private-role-panel.component';
import { RoleRevealComponent } from './core/game/role-reveal/role-reveal.component';
import { GameApiService } from './shared/api/game/game-api.service';
import { AuthTokenInterceptor } from './shared/api/auth/auth-token.interceptor';

@NgModule({
  declarations: [
    AppComponent,
    GameComponent,
    ForensicComponent,
    AllGamesComponent,
    CardComponent,
    ActiveGamesListItemComponent,
    JoinGameComponent,
    JoinedPlayersListComponent,
    JoinedPlayerListItemComponent,
    RippleMatCardComponent,
    PlayerDeckComponent,
    ChatComponent,
    PlayerDeckPagerComponent,
    GuessComponent,
    ForensicCardComponent,
    MurdererSelectDialogComponent,
    TestComponent,
    CenteringFlexComponent,
    AvatarComponent,
    NavbarComponent,
    CopyLinkComponent,
    InfoComponent,
    MurdererInfoComponent,
    ConfirmationButtonComponent,
    PromptPanelComponent,
    WitnessSelectionPromptComponent,
    ModeratorWitnessControlsComponent,
    PrivateRolePanelComponent,
    RoleRevealComponent
  ],
  imports: [
    BrowserModule,
    AppRoutingModule,
    BrowserAnimationsModule,
    HttpClientModule,
    FormsModule,
    MatButtonModule,
    MatDividerModule,
    MatInputModule,
    MatToolbarModule,
    MatProgressSpinnerModule,
    MatRippleModule,
    MatListModule,
    MatDialogModule,
    MatCardModule,
    MatProgressBarModule,
    LazyLoadImageModule,
    MatSnackBarModule
  ],
  providers: [
    GameApiService,
    { provide: HTTP_INTERCEPTORS, useClass: AuthTokenInterceptor, multi: true },
    { provide: MAT_DIALOG_DEFAULT_OPTIONS, useValue: { hasBackdrop: false } }
  ],
  bootstrap: [AppComponent],
  entryComponents: [MurdererSelectDialogComponent]
})
export class AppModule {}
