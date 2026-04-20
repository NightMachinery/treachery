import { ForensicCardComponent } from './forensic-card.component';

describe('ForensicCardComponent', () => {
  function createComponent() {
    const component = new ForensicCardComponent();
    component.forensicCard = {
      cardName: 'Locations',
      choices: ['Living Room', 'Bedroom'],
      selectedChoice: '',
      replaced: false
    };
    return component;
  }

  it('does not highlight the transient selected option for cards that are not selected', () => {
    const component = createComponent();
    component.selected = false;
    component.selectedOptionName = 'Bedroom';

    expect(component.isSelected('Bedroom')).toBe(false);
  });

  it('highlights the transient selected option for the selected card', () => {
    const component = createComponent();
    component.selected = true;
    component.selectedOptionName = 'Bedroom';

    expect(component.isSelected('Bedroom')).toBe(true);
  });

  it('always highlights persisted selected choices', () => {
    const component = createComponent();
    component.selected = false;
    component.forensicCard.selectedChoice = 'Living Room';

    expect(component.isSelected('Living Room')).toBe(true);
  });
});
