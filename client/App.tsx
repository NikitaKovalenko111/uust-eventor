import { StatusBar } from 'expo-status-bar';
import { Provider } from 'react-redux';
import { RootNavigator } from './src/navigation/RootNavigator';
import { store } from './src/redux/store';

export default function App() {
  return (
    <Provider store={store}>
      <StatusBar style="dark" />
      <RootNavigator />
    </Provider>
  );
}
