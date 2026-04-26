import { RouterProvider, createRouter } from '@tanstack/react-router'
import { useAuth } from './contexts/AuthContext'
import { routeTree } from './routeTree.gen'

// Define the router context type test
interface RouterContext {
  auth: {
    isAuthenticated: boolean
    isLoading: boolean
  }
}

// Create the router instance
const router = createRouter({
  routeTree,
  context: {
    auth: {
      isAuthenticated: false,
      isLoading: true,
    },
  } as RouterContext,
})

// Register the router for type safety
declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}

function App() {
  const auth = useAuth()
  
  return <RouterProvider router={router} context={{ auth }} />
}

export default App
