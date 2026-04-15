import { of } from 'rxjs';

import { ForensicComponent } from './forensic.component';

describe('ForensicComponent', () => {
  function buildGame(selectedCount: number) {
    return {
      causeCard: { cardName: 'Cause', choices: ['A'], selectedChoice: 'A', replaced: false },
      locationCard: { cardName: 'Location', choices: ['B'], selectedChoice: 'B', replaced: false },
      otherCards: Array.from({ length: 6 }, (_, index) => ({
        cardName: `Other ${index + 1}`,
        choices: ['One', 'Two'],
        selectedChoice: index < selectedCount ? 'One' : '',
        replaced: false
      }))
    } as any;
  }

  function createComponent(game: any) {
    const gameApi = {
      game$: of(game),
      selectNextForensicOtherCard: jasmine.createSpy('selectNextForensicOtherCard')
    };
    const snack = {
      error: jasmine.createSpy('error')
    };

    const component = new ForensicComponent(
      { params: of({ gameId: 'ABCD' }) } as any,
      {} as any,
      {} as any,
      gameApi as any,
      {} as any,
      { user: { uid: 'creator' } } as any,
      {} as any,
      snack as any
    );

    return { component, gameApi, snack };
  }

  it('blocks selecting the next other card when a replacement is required but missing', async () => {
    const game = buildGame(4);
    const { component, gameApi, snack } = createComponent(game);
    component.selectedOtherCardOption = 'Two';

    await component.selectNextOtherCard();

    expect(snack.error).toHaveBeenCalledWith('Please select a card to replace first!');
    expect(gameApi.selectNextForensicOtherCard).not.toHaveBeenCalled();
  });

  it('sends replaceCardName when selecting the next other card after choosing a replacement', async () => {
    const game = buildGame(4);
    const { component, gameApi, snack } = createComponent(game);
    component.selectedOtherCardOption = 'Two';
    component.replaceCardName = 'Other 2';

    await component.selectNextOtherCard();

    expect(snack.error).not.toHaveBeenCalled();
    expect(gameApi.selectNextForensicOtherCard).toHaveBeenCalledWith(
      jasmine.objectContaining({
        cardName: 'Other 5',
        selectedChoice: 'Two'
      }),
      'Other 2'
    );
  });
});
