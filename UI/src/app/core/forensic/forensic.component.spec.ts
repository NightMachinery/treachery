import { of } from 'rxjs';

import { TgForensicCard } from '../../shared/api/models/models';
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

  function buildLocationCards(): TgForensicCard[] {
    return [
      { cardName: 'Locations', choices: ['Living Room', 'Bedroom'], selectedChoice: '', replaced: false },
      { cardName: 'Locations', choices: ['Vacation Home', 'Park'], selectedChoice: '', replaced: false },
      { cardName: 'Locations', choices: ['Pub', 'Hotel'], selectedChoice: '', replaced: false }
    ];
  }

  function createComponent(game: any, overrides: { gameApi?: any; snack?: any } = {}) {
    const gameApi = overrides.gameApi || {
      game$: of(game),
      selectNextForensicOtherCard: jasmine.createSpy('selectNextForensicOtherCard'),
      selectForensicLocationCard: jasmine.createSpy('selectForensicLocationCard')
    };
    const snack = overrides.snack || {
      error: jasmine.createSpy('error')
    };

    const component = new ForensicComponent(
      {
        params: of({ gameId: 'ABCD' }),
        snapshot: { queryParamMap: { get: () => null } },
        queryParams: of({})
      } as any,
      {} as any,
      {} as any,
      gameApi as any,
      {} as any,
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

  it('sends the exact selected location card when duplicate card names exist', async () => {
    const locationCards = buildLocationCards();
    const { component, gameApi } = createComponent(buildGame(0));
    const selectedCard = locationCards[2];

    component.locationCardClick(selectedCard);
    component.selectedLocationCardOption = 'Hotel';

    await component.selectLocationCard();

    expect(gameApi.selectForensicLocationCard).toHaveBeenCalledTimes(1);
    expect(gameApi.selectForensicLocationCard).toHaveBeenCalledWith({
      ...selectedCard,
      selectedChoice: 'Hotel'
    });
    expect(component.selectedLocationCard).toBeNull();
    expect(component.selectedLocationCardOption).toBeNull();
  });

  it('resets the location option to the clicked card first choice when switching cards with the same name', () => {
    const [firstCard, secondCard] = buildLocationCards();
    const { component } = createComponent(buildGame(0));

    component.locationCardClick(firstCard);
    component.selectedLocationCardOption = firstCard.choices[1];

    component.locationCardClick(secondCard);

    expect(component.selectedLocationCard).toBe(secondCard);
    expect(component.selectedLocationCardOption).toBe(secondCard.choices[0]);
  });
});
