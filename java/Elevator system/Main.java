import java.util.*;

/*  
requirements:
Request: floor, direction
    inside, select destination floor.
Elevator: tracks its states:
    - current floor
    - direction
    - state : IDLE, MOVING, MAINTENANCE
  stops at requested floors
ElevatorSystem: dispatches available elevators
Maintainer: can take Elevator out of service and return to service.

To implement: request, dispatch, movement logic
decide how to assign elevators to requests.

*/

enum Direction {UP, DOWN}
enum State {IDLE, MOVING, MAINTENANCE}

class Elevator {
    final int id;
    int currentFloor = 0;
    Direction direction = Direction.UP;
    State state = State.IDLE;
    PriorityQueue<Integer> minq = new PriorityQueue<>();;
    PriorityQueue<Integer> maxq = new PriorityQueue<>(Collections.reverseOrder());

    public Elevator(int id){
        this.id = id;
    }

    void addStops(int floor){
        if(direction == Direction.UP){
            if(floor > currentFloor){
                minq.add(floor);
            } else {
                maxq.add(floor);
            }
        } else {
            if(floor < currentFloor){
                maxq.add(floor);
            } else {
                minq.add(floor);
            }
        }
    }

    void addDestination(int floor){
        if(state == State.MAINTENANCE) return;
        addStops(floor);
    }

    public void setMaintenance(boolean maintenance){
        if(maintenance){
            state = State.MAINTENANCE;
        } else {
            state = State.IDLE;
        }
    }

    public void dispatch(int floor){
        if(state == State.MAINTENANCE) return;
        this.direction = (floor > currentFloor) ? Direction.UP : Direction.DOWN;
        addDestination(floor);
        state = State.MOVING;
    }

    public void move(){
        if(state == State.MAINTENANCE && minq.isEmpty() && maxq.isEmpty()) return;
        if(direction == Direction.UP){
            if(!minq.isEmpty()){
                int nextFloor = minq.peek();
                if(nextFloor > currentFloor){
                    currentFloor++;
                } else if(nextFloor == currentFloor){
                    minq.poll();
                }
            } else if(!maxq.isEmpty()) {
                direction = Direction.DOWN;
            } else {
                if(state != State.MAINTENANCE) {
                    state = State.IDLE;
                }
            }
        } else {
            if(!maxq.isEmpty()){
                int nextFloor = maxq.peek();
                if(nextFloor < currentFloor){
                    currentFloor--;
                } else if(nextFloor == currentFloor){
                    maxq.poll();
                }
            } else if(!minq.isEmpty()){
                direction = Direction.UP;
            } else {
                if(state != State.MAINTENANCE) {
                    state = State.IDLE;
                }
            }
        }
    }
}

class ElevatorSystem {
    List<Elevator> elevators = new ArrayList<>();

    ElevatorSystem(int numElevators){
        for(int i=0; i<numElevators; i++){
            elevators.add(new Elevator(i));
        }
    }

    public void request(int floor, Direction dir){
        for(Elevator e : elevators){
            if(e.state == State.MOVING && e.direction == dir){
                e.addDestination(floor);
                return;
            }
        }
        for(Elevator e : elevators){
            if(e.state == State.IDLE){
                e.dispatch(floor);
                return;
            }
        }
        for(Elevator e : elevators){
            if(e.state == State.MOVING){
                e.addDestination(floor);
                return;
            }
        }
    }
}

public class Main {
    public static void main(String[] args){
        ElevatorSystem system = new ElevatorSystem(3);
        system.request(5, Direction.UP);
        system.request(2, Direction.DOWN);
        system.request(8, Direction.UP);
        system.request(1, Direction.DOWN);

        for(int i=0; i<10; i++){
            for(Elevator e : system.elevators){
                e.move();
                System.out.println("Elevator " + e.id + " is at floor " + e.currentFloor + " moving " + e.direction + " state: " + e.state);
            }
            System.out.println("-----");
        }
    }
}
