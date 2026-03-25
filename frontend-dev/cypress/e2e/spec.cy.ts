/// <reference types="cypress" />
describe('Stock test', () => {
  it('Looks for Stock Page', () => {
    // go to stock page
    cy.visit("http://localhost:4200/stocks")
    
    // type Google ticker into search bar
    cy.get("input.search-bar")
    .type("GOOG")

    // Click search button
    cy.contains("Search stock")
    .click()

    // Check that ticker is updated after search
    cy.get("div.card-area")
    .contains("GOOG")
    
    // Check price is updated after search
    cy.get("div.card-area")
    .contains("298.75")

    // Check that chart is opened after search
    cy.get("div.chart")
    .click()
    .contains("data.open")
  })
})