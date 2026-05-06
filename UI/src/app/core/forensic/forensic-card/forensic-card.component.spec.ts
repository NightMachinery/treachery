import { ForensicCardComponent } from './forensic-card.component';

describe('ForensicCardComponent', () => {
  function createComponent() {
    const component = new ForensicCardComponent();
    component.forensicCard = {
      cardId: 'locations-1',
      cardName: 'Locations',
      choices: ['Living Room', 'Bedroom'],
      choiceIds: ['living-room', 'bedroom'],
      selectedChoiceId: '',
      selectedChoice: '',
      replaced: false,
    };
    return component;
  }

  it('does not highlight the transient selected option for cards that are not selected', () => {
    const component = createComponent();
    component.selected = false;
    component.selectedOptionId = 'bedroom';

    expect(component.isSelected('bedroom')).toBe(false);
  });

  it('highlights the transient selected option for the selected card', () => {
    const component = createComponent();
    component.selected = true;
    component.selectedOptionId = 'bedroom';

    expect(component.isSelected('bedroom')).toBe(true);
  });

  it('always highlights persisted selected choices', () => {
    const component = createComponent();
    component.selected = false;
    component.forensicCard.selectedChoiceId = 'living-room';
    component.forensicCard.selectedChoice = 'Living Room';

    expect(component.isSelected('living-room')).toBe(true);
  });
});
