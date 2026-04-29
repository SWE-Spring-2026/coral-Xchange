import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { of } from 'rxjs';
import { SignUpPage } from './sign-up-page';
import { Auth } from '../auth';

type RegisterPayload = {
  name: string;
  username: string;
  email: string;
  password: string;
};

describe('SignUpPage', () => {
  let component: SignUpPage;
  let fixture: ComponentFixture<SignUpPage>;
  let registerCalls: RegisterPayload[];

  const authMock = {
    register: (payload: RegisterPayload) => {
      registerCalls.push(payload);
      return of({});
    },
  };

  beforeEach(async () => {
    registerCalls = [];
    await TestBed.configureTestingModule({
      imports: [SignUpPage],
      providers: [provideRouter([]), { provide: Auth, useValue: authMock }],
    })
    .compileComponents();

    fixture = TestBed.createComponent(SignUpPage);
    component = fixture.componentInstance;
    fixture.detectChanges();
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should start with an invalid form', () => {
    expect(component.signUpForm.invalid).toBeTruthy();
  });

  it('should keep optional fields optional', () => {
    component.signUpForm.patchValue({
      name: 'Carlos Diaz',
      username: 'carlos',
      email: 'carlos@example.com',
      password: 'password123',
      confirmPassword: 'password123',
      tradingExperience: '',
      interestedSectors: [],
    });

    expect(component.signUpForm.valid).toBeTruthy();
  });

  it('should be invalid when passwords do not match', () => {
    component.signUpForm.patchValue({
      name: 'Carlos Diaz',
      username: 'carlos',
      email: 'carlos@example.com',
      password: 'password123',
      confirmPassword: 'different123',
    });

    expect(component.signUpForm.invalid).toBeTruthy();
    expect(component.signUpForm.hasError('passwordMismatch')).toBeTruthy();
  });

  it('should submit only the backend-supported signup fields', () => {
    component.signUpForm.patchValue({
      name: 'Carlos Diaz',
      username: 'carlos',
      email: 'carlos@example.com',
      password: 'password123',
      confirmPassword: 'password123',
      tradingExperience: 'Beginner',
      interestedSectors: ['Technology', 'Finance'],
    });

    component.submit();

    expect(registerCalls.length).toBe(1);
    expect(registerCalls[0]).toEqual({
      name: 'Carlos Diaz',
      username: 'carlos',
      email: 'carlos@example.com',
      password: 'password123',
    });
  });
});
