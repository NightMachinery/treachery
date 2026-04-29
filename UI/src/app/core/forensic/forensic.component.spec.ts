import { of } from 'rxjs';

import { TgForensicCard } from '../../shared/api/models/models';
import { ForensicComponent } from './forensic.component';

describe('ForensicComponent', () => {
  function buildGame(selectedCount: number) {
    return {
      causeCard: {
        cardId: 'cause',
        cardName: 'Cause',
        choices: ['A'],
        choiceIds: ['a'],
        selectedChoiceId: 'a',
        selectedChoice: 'A',
        replaced: false,
      },
      locationCard: {
        cardId: 'location',
        cardName: 'Location',
        choices: ['B'],
        choiceIds: ['b'],
        selectedChoiceId: 'b',
        selectedChoice: 'B',
        replaced: false,
      },
      otherCards: Array.from({ length: 6 }, (_, index) => ({
        cardId: `other-${index + 1}`,
        cardName: `Other ${index + 1}`,
        choices: ['One', 'Two'],
        choiceIds: ['one', 'two'],
        selectedChoiceId: index < selectedCount ? 'one' : '',
        selectedChoice: index < selectedCount ? 'One' : '',
        replaced: false,
      })),
    } as any;
  }

  function buildLocationCards(): TgForensicCard[] {
    return [
      {
        cardId: 'locations-1',
        cardName: 'Locations',
        choices: ['Living Room', 'Bedroom'],
        choiceIds: ['living-room', 'bedroom'],
        selectedChoiceId: '',
        selectedChoice: '',
        replaced: false,
      },
      {
        cardId: 'locations-2',
        cardName: 'Locations',
        choices: ['Vacation Home', 'Park'],
        choiceIds: ['vacation-home', 'park'],
        selectedChoiceId: '',
        selectedChoice: '',
        replaced: false,
      },
      {
        cardId: 'locations-3',
        cardName: 'Locations',
        choices: ['Pub', 'Hotel'],
        choiceIds: ['pub', 'hotel'],
        selectedChoiceId: '',
        selectedChoice: '',
        replaced: false,
      },
    ];
  }

  function buildCauseCards(): TgForensicCard[] {
    return [
      {
        cardId: 'causes-1',
        cardName: 'Cause of Death',
        choices: ['Suffocation', 'Blood Loss'],
        choiceIds: ['suffocation', 'blood-loss'],
        selectedChoiceId: '',
        selectedChoice: '',
        replaced: false,
      },
      {
        cardId: 'causes-2',
        cardName: 'Cause of Death',
        choices: ['Illness', 'Poisoning'],
        choiceIds: ['illness', 'poisoning'],
        selectedChoiceId: '',
        selectedChoice: '',
        replaced: false,
      },
    ];
  }

  function createComponent(game: any, overrides: { gameApi?: any; snack?: any } = {}) {
    const gameApi = overrides.gameApi || {
      game$: of(game),
      selectNextForensicOtherCard: jasmine.createSpy('selectNextForensicOtherCard'),
      selectForensicLocationCard: jasmine.createSpy('selectForensicLocationCard'),
    };
    const snack = overrides.snack || {
      error: jasmine.createSpy('error'),
    };

    const component = new ForensicComponent(
      {
        params: of({ gameId: 'ABCD' }),
        snapshot: { queryParamMap: { get: () => null } },
        queryParams: of({}),
      } as any,
      {} as any,
      {} as any,
      gameApi as any,
      {} as any,
      {} as any,
      snack as any,
    );

    return { component, gameApi, snack };
  }

  it('blocks selecting the next other card when a replacement is required but missing', async () => {
    const game = buildGame(4);
    const { component, gameApi, snack } = createComponent(game);
    component.selectedOtherCardOptionId = 'two';

    await component.selectNextOtherCard();

    expect(snack.error).toHaveBeenCalledWith('Please select a card to replace first!');
    expect(gameApi.selectNextForensicOtherCard).not.toHaveBeenCalled();
  });

  it('sends replaceCardId when selecting the next other card after choosing a replacement', async () => {
    const game = buildGame(4);
    const { component, gameApi, snack } = createComponent(game);
    component.selectedOtherCardOptionId = 'two';
    component.replaceCardId = 'other-2';

    await component.selectNextOtherCard();

    expect(snack.error).not.toHaveBeenCalled();
    expect(gameApi.selectNextForensicOtherCard).toHaveBeenCalledWith(
      jasmine.objectContaining({
        cardId: 'other-5',
        cardName: 'Other 5',
        selectedChoiceId: 'two',
        selectedChoice: 'Two',
      }),
      'other-2',
    );
  });

  it('sends the exact selected location card when duplicate card names exist', async () => {
    const locationCards = buildLocationCards();
    const { component, gameApi } = createComponent(buildGame(0));
    const selectedCard = locationCards[2];

    component.locationCardClick(selectedCard);
    component.selectedLocationCardOptionId = 'hotel';

    await component.selectLocationCard();

    expect(gameApi.selectForensicLocationCard).toHaveBeenCalledTimes(1);
    expect(gameApi.selectForensicLocationCard).toHaveBeenCalledWith({
      ...selectedCard,
      selectedChoiceId: 'hotel',
      selectedChoice: 'Hotel',
    });
    expect(component.selectedLocationCard).toBeNull();
    expect(component.selectedLocationCardOptionId).toBeNull();
  });

  it('defaults the cause option to the clicked card first choice', () => {
    const [firstCard] = buildCauseCards();
    const { component } = createComponent(buildGame(0));

    component.causeCardClick(firstCard);

    expect(component.selectedCauseCardId).toBe(firstCard.cardId);
    expect(component.selectedCauseCardOptionId).toBe(firstCard.choiceIds[0]);
  });

  it('preserves the chosen cause option when clicking inside the already selected card', () => {
    const [firstCard] = buildCauseCards();
    const { component } = createComponent(buildGame(0));

    component.causeCardClick(firstCard);
    component.selectedCauseCardOptionId = firstCard.choiceIds[1];

    component.causeCardClick(firstCard);

    expect(component.selectedCauseCardId).toBe(firstCard.cardId);
    expect(component.selectedCauseCardOptionId).toBe(firstCard.choiceIds[1]);
  });

  it('resets the cause option to the clicked card first choice when switching cards', () => {
    const [firstCard, secondCard] = buildCauseCards();
    const { component } = createComponent(buildGame(0));

    component.causeCardClick(firstCard);
    component.selectedCauseCardOptionId = firstCard.choiceIds[1];

    component.causeCardClick(secondCard);

    expect(component.selectedCauseCardId).toBe(secondCard.cardId);
    expect(component.selectedCauseCardOptionId).toBe(secondCard.choiceIds[0]);
  });

  it('resets the location option to the clicked card first choice when switching cards with the same name', () => {
    const [firstCard, secondCard] = buildLocationCards();
    const { component } = createComponent(buildGame(0));

    component.locationCardClick(firstCard);
    component.selectedLocationCardOptionId = firstCard.choiceIds[1];

    component.locationCardClick(secondCard);

    expect(component.selectedLocationCard).toBe(secondCard);
    expect(component.selectedLocationCardOptionId).toBe(secondCard.choiceIds[0]);
  });
});
