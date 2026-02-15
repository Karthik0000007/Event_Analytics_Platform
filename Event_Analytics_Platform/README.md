# Event Analytics Platform — Frontend Documentation

## Overview

The frontend application of the Event Analytics Platform is built using React and TypeScript. It serves as the user interface for interacting with the backend event analytics services. This documentation provides an overview of the project structure, setup instructions, and usage guidelines.

## Project Structure

```
frontend/
├── package.json        # Configuration file for npm, listing dependencies and scripts
├── tsconfig.json       # TypeScript configuration file specifying compiler options
├── vite.config.ts      # Configuration file for Vite, the build tool
├── public/
│   └── index.html      # Main HTML file serving as the entry point for the application
├── src/
│   ├── main.tsx        # Entry point of the React application
│   ├── App.tsx         # Main application component
│   ├── components/
│   │   └── Header.tsx  # Reusable header component
│   ├── pages/
│   │   └── Home.tsx    # Home page component
│   ├── hooks/
│   │   └── useApi.ts    # Custom hook for handling API calls
│   ├── services/
│   │   └── api.ts      # Functions for making API requests to the backend
│   ├── styles/
│   │   └── index.css    # Global CSS styles for the application
│   └── types/
│       └── index.ts     # TypeScript interfaces and types for type safety
```

## Setup Instructions

1. **Install Dependencies**: Navigate to the `frontend` directory and run the following command to install the necessary dependencies:

   ```
   npm install
   ```

2. **Run the Development Server**: Start the development server using Vite:

   ```
   npm run dev
   ```

   This will start the application and make it accessible at `http://localhost:3000`.

3. **Build for Production**: To create a production build of the application, run:

   ```
   npm run build
   ```

   The built files will be generated in the `dist` directory.

## Usage Guidelines

- The application is structured to facilitate easy navigation and component reuse.
- Use the `Header` component in other pages to maintain a consistent layout.
- The `useApi` hook can be utilized in any component to fetch data from the backend.
- Ensure to define TypeScript interfaces in the `types` directory for better type safety across the application.

## Contributing

Contributions to the frontend application are welcome. Please follow the standard practices for code contributions, including writing tests and updating documentation as necessary.

## License

This project is licensed under the MIT License. See the LICENSE file for more details.