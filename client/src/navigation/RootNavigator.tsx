import { NavigationContainer } from '@react-navigation/native';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { useEffect } from 'react';
import { useAppSelector } from '../redux/hooks';
import { setApiAccessToken } from '../api/api';
import { AuthPage } from '../pages/AuthPage';
import { UserCabinetPage } from '../pages/UserCabinetPage';
import { ModeratorCabinetPage } from '../pages/ModeratorCabinetPage';
import { EventsPage } from '../pages/EventsPage';
import { EventDetailsPage } from '../pages/EventDetailsPage';
import { RootStackParamList } from './types';

const Stack = createNativeStackNavigator<RootStackParamList>();

export const RootNavigator = () => {
  const isAuthenticated = useAppSelector((state) => state.auth.isAuthenticated);
  const token = useAppSelector((state) => state.auth.token);
  const userRole = useAppSelector((state) => state.auth.user?.role);

  useEffect(() => {
    setApiAccessToken(token);
  }, [token]);

  return (
    <NavigationContainer>
      <Stack.Navigator
        initialRouteName={isAuthenticated ? undefined : 'Events'}
        screenOptions={{ headerShown: false, animation: 'slide_from_right' }}
      >
        {!isAuthenticated ? (
          <>
            <Stack.Screen name="Events" component={EventsPage} />
            <Stack.Screen name="EventDetails" component={EventDetailsPage} />
            <Stack.Screen name="Auth" component={AuthPage} />
          </>
        ) : (
          <>
            {userRole === 'moderator' ? (
              <Stack.Screen name="ModeratorCabinet" component={ModeratorCabinetPage} />
            ) : (
              <Stack.Screen name="UserCabinet" component={UserCabinetPage} />
            )}
            <Stack.Screen name="Events" component={EventsPage} />
            <Stack.Screen name="EventDetails" component={EventDetailsPage} />
          </>
        )}
      </Stack.Navigator>
    </NavigationContainer>
  );
};
