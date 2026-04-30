/// <reference types="cypress" />
describe('Stock test', () => {
  it('Looks for Stock Page', () => {
    // go to stock page
    cy.visit("http://localhost:4200/stocks")
    
    // type Google ticker into search bar
    cy.get("input.search-input")
    .type("GOOG")

    // Click search button
    cy.contains("Search")
    .click()
  })
})