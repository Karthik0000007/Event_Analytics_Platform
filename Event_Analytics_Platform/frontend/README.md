# Frontend Application

This directory contains the frontend application for the Event Analytics Platform. The application is built using React and TypeScript, and it serves as the user interface for interacting with the backend event analytics services.

## Project Structure

- **package.json**: Configuration file for npm, listing dependencies and scripts.
- **tsconfig.json**: TypeScript configuration file specifying compiler options.
- **vite.config.ts**: Configuration file for Vite, the build tool used for serving the application.
- **public/index.html**: Main HTML file that serves as the entry point for the application.
- **src/**: Contains the source code for the application.
  - **main.tsx**: Entry point of the React application.
  - **App.tsx**: Main application component that sets up routing and includes other components.
  - **components/**: Contains reusable components.
    - **Header.tsx**: Header component for the application.
  - **pages/**: Contains page components.
    - **Home.tsx**: Home page component.
  - **hooks/**: Contains custom hooks.
    - **useApi.ts**: Custom hook for handling API calls.
  - **services/**: Contains functions for making API requests.
    - **api.ts**: API request functions.
  - **styles/**: Contains global CSS styles.
    - **index.css**: Global styles for the application.
  - **types/**: Contains TypeScript interfaces and types.
    - **index.ts**: Type definitions used throughout the application.

## Setup Instructions

1. **Install Dependencies**: Run `npm install` to install the required dependencies.
2. **Run the Application**: Use `npm run dev` to start the development server.
3. **Build for Production**: Use `npm run build` to create a production build of the application.

## Usage Guidelines

- The application is designed to interact with the backend Event Analytics Platform.
- Ensure that the backend services are running before starting the frontend application.
- Modify the components and pages as needed to fit the requirements of your application.

For more detailed information on each file and component, refer to the respective files in the `src` directory.