import { NavigationContainer } from '@react-navigation/native';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { useEffect } from 'react';
import { useAppDispatch, useAppSelector } from '../redux/hooks';
import { fetchEventsRequest } from '../redux/slices/eventsSlice';
import { AuthPage } from '../pages/AuthPage';
import { UserCabinetPage } from '../pages/UserCabinetPage';
import { ModeratorCabinetPage } from '../pages/ModeratorCabinetPage';
import { EventsPage } from '../pages/EventsPage';
import { EventDetailsPage } from '../pages/EventDetailsPage';
import { RootStackParamList } from './types';

const Stack = createNativeStackNavigator<RootStackParamList>();

export const RootNavigator = () => {
  const dispatch = useAppDispatch();
  const isAuthenticated = useAppSelector((state) => state.auth.isAuthenticated);
  const userRole = useAppSelector((state) => state.auth.user?.role);

  useEffect(() => {
    if (isAuthenticated) {
      dispatch(fetchEventsRequest());
    }
  }, [dispatch, isAuthenticated]);

  return (
    <NavigationContainer>
      <Stack.Navigator screenOptions={{ headerShown: false, animation: 'slide_from_right' }}>
        {!isAuthenticated ? (
          <Stack.Screen name="Auth" component={AuthPage} />
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
